package server

import (
	"net"
	"strconv"
	"sync"

	"github.com/qyqx233/go-tunel/lib"

	"go.uber.org/zap"
)

func NewTcpConn(host string, port int) (conn net.Conn, er error) {
	addr, err := net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		return
	}
	conn, err = net.DialTCP("tcp", nil, addr)
	return
}

var logger *zap.SugaredLogger

func init() {
	logger = lib.GetLogger()
}

type HandleFunc = func(net.Conn, []byte)

type TransportServerStru struct {
	conn   net.Conn
	addr   string
	host   string
	port   int
	Handle HandleFunc
}

func With(func(transport *TransportServerStru)) {

}

func NewTransportServer(addr string, host string, port int) *TransportServerStru {
	t := TransportServerStru{nil, addr, host, port, nil}
	return &t
}

func NewTransportDebugServer(addr string, host string, port int) TransportServerStru {
	t := TransportServerStru{nil, addr, host, port, nil}
	return t
}

func (t *TransportServerStru) TransportExt(conn net.Conn) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	conn1, err := NewTcpConn(t.host, t.port)
	if err != nil {
		logger.Error(err)
		conn.Close()
		return
	}
	wc := lib.NewWrapConn(conn, lib.NextUid())
	wc1 := lib.NewWrapConn(conn1, lib.NextUid())
	go lib.Pipe2(wg, wc, wc1, func() {
		conn.Close()
		conn1.Close()
	})
	lib.Pipe3(wg, wc1, wc, func() {
		conn.Close()
		conn1.Close()
	}, t.Handle)
	logger.Info("链接断开")
	wg.Wait()
}

func (t *TransportServerStru) Transport(conn net.Conn) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	conn1, err := NewTcpConn(t.host, t.port)
	if err != nil {
		logger.Error(err)
		conn.Close()
		return
	}
	wc := lib.NewWrapConn(conn, lib.NextUid())
	wc1 := lib.NewWrapConn(conn1, lib.NextUid())
	if t.Handle == nil {
		go lib.Pipe2(wg, wc, wc1, func() {
			conn.Close()
			conn1.Close()
		})
		go lib.Pipe2(wg, wc1, wc, func() {
			conn.Close()
			conn1.Close()
		})
	} else {
		go lib.Pipe3(wg, wc, wc1, func() {
			conn.Close()
			conn1.Close()
		}, t.Handle)
		go lib.Pipe3(wg, wc1, wc, func() {
			conn.Close()
			conn1.Close()
		}, t.Handle)
	}
	logger.Info("链接断开")
	wg.Wait()
}

func (t TransportServerStru) Shutdown() {
	t.conn.Close()
}

func (t TransportServerStru) Start() error {
	listener, err := net.Listen("tcp", t.addr)
	if err != nil {
		logger.Error(err)
		return err
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			logger.Info("客户端地址:" + conn.RemoteAddr().String())
			go t.Transport(conn)
		}
	}()
	return nil
}
