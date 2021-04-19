package lib

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qyqx233/go-tunel/lib/proto"
)

type LogConn interface {
	net.Conn
	ID() int64
}

var maxUint63 uint64 = 2<<62 - 1
var uid uint64

func NextUid() int64 {
	return int64(atomic.AddUint64(&uid, 1) & maxUint63)
}

func NextPosUid() int64 {
	return int64(atomic.AddUint64(&uid, 1)&maxUint63) * -1
}

type WrapConnStru struct {
	net.Conn
	id     int64
	atomic int32
}

func (w WrapConnStru) ID() int64 {
	return w.id
}

func (w *WrapConnStru) ShutDown() error {
	if atomic.CompareAndSwapInt32(&w.atomic, 0, 1) {
		logger.Infof("通道%d被关闭", w.ID())
		return w.Conn.Close()
	}
	return nil
}

// outer端调用生成正的int64
func NewWrapConn(c net.Conn, id int64) WrapConnStru {
	if id == 0 {
		return WrapConnStru{c, int64(atomic.AddUint64(&uid, 1) & maxUint63), 0}
	} else {
		return WrapConnStru{c, id, 0}
	}
}

func Pipe(wg *sync.WaitGroup, to net.Conn, from net.Conn) {
	var err error
	n, err := io.Copy(to, from)
	if err != nil {
		logger.Errorf("io.Copy err = %v", err)
	} else {
		logger.Infof("io.Copy %d bytes", n)
	}
	from.Close()
	to.Close()
	wg.Done()
}

type rwError struct {
	error
	d string
}

var ioCopy = io.Copy

func Copy(dst net.Conn, src net.Conn, h func(net.Conn, []byte)) (written int64, err error) {
	return copyBuffer(dst, src, nil, h)
}

func Copy2(dst net.Conn, src net.Conn) (written int64, err error) {
	d := dst.(*net.TCPConn)
	s := dst.(*net.TCPConn)
	return d.ReadFrom(s)
}

func copyBuffer(dst net.Conn, src net.Conn, buf []byte, h func(net.Conn, []byte)) (written int64, err error) {
	// if wt, ok := src.(io.WriterTo); ok {
	// logger.Debug("WriterTo")
	// return wt.WriteTo(dst)
	// }
	// if rt, ok := dst.(io.ReaderFrom); ok {
	// return rt.ReadFrom(src)
	// }
	if buf == nil {
		size := 32 * 1024
		// if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
		// 	if l.N < 1 {
		// 		size = 1
		// 	} else {
		// 		size = int(l.N)
		// 	}
		// }
		buf = make([]byte, size)
	}
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			h(src, buf[:nr])
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = rwError{ew, "w"}
				break
			}
			if nr != nw {
				err = rwError{io.ErrShortWrite, "w"}
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = rwError{er, "r"}
			}
			break
		}
	}
	return written, err
}
func Pipe2(wg *sync.WaitGroup, to WrapConnStru, from WrapConnStru, done func()) {
	var err error
	var n int64
	n, err = io.Copy(to.Conn, from.Conn)
	for {
		if err != nil {
			logger.Errorf("%d->%d io.Copy failed: %v", from.ID(), to.ID(), err)
			break
		} else {
			logger.Infof("%d->%d io.Copy %d bytes", from.ID(), to.ID(), n)
			break
		}
	}
	done()
	wg.Done()
}

func Pipe3(wg *sync.WaitGroup, to WrapConnStru, from WrapConnStru, done func(), h func(net.Conn, []byte)) {
	var err error
	var n int64
	n, err = Copy(to.Conn, from.Conn, h)
	for {
		if err != nil {
			logger.Errorf("%d->%d io.Copy failed: %v", from.ID(), to.ID(), err)
			break
		} else {
			logger.Infof("%d->%d io.Copy %d bytes", from.ID(), to.ID(), n)
			break
		}
	}
	done()
	wg.Done()
}

func SetTcpOptions(conn net.Conn, options ...int) {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return
	}
	// for option := range options {
	// 	switch option {
	// 	case proto.KeepAliveTcpOption:
	// 		tcpConn.SetKeepAlive(true)
	// 	}
	// }
	for i := 0; i < len(options)/2; i++ {
		k, v := options[2*i], options[2*i+1]
		switch k {
		case proto.NoDelayOption:
			if v == proto.True {
				tcpConn.SetNoDelay(true)
			} else {
				tcpConn.SetNoDelay(false)
			}
		case proto.KeepAliveTcpOption:
			if v == proto.True {
				tcpConn.SetKeepAlive(true)
			} else {
				tcpConn.SetKeepAlive(true)
			}
		case proto.AlivePeriodOption:
			tcpConn.SetKeepAlivePeriod(time.Duration(int64(time.Second) * int64(v)))
		}
	}
}
