package inner

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"gotunel/lib"
	"gotunel/lib/proto"
)

var maxUint63 uint64 = 2<<62 - 1

type transport struct {
	minConns   int
	maxConns   int
	targetPort int
	// targetHost [32]byte
	targetHost string
	tcpOrUdp   string
	serverIp   string
	serverPort int
	name       [16]byte
	symkey     [16]byte
	idleConns  int32
	connCh     chan net.Conn
	conns      []net.Conn
	atomic     int32
}

func (t *transport) shake(conn net.Conn, transportType int8, usage int8, reqID int64, corrReqID int64) error {
	shake := proto.ShakeProto{
		Type:      transportType,
		Usage:     usage,
		Name:      t.name,
		SymKey:    t.symkey,
		Host:      lib.String2Byte32(t.targetHost),
		Port:      t.targetPort,
		ReqID:     reqID,
		CorrReqId: corrReqID,
	}
	err := shake.Send(conn)
	if err != nil {
		logger.Error(err)
		return err
	}
	if transportType == proto.CmdType {
		err = shake.Recv(conn)
		if err != nil {
			logger.Error(err)
			return err
		}
		if shake.Code != proto.OkCode {
			logger.Errorf("握手返回错误码%d", shake.Code)
			return errors.New("握手失败")
		}
	}
	return nil
}

// 使用go xxx调用
func (t *transport) createCmdAndConn() error {
	if !atomic.CompareAndSwapInt32(&t.atomic, 0, 1) {
		return nil
	}
	addr, _ := net.ResolveTCPAddr("tcp", t.serverIp+":"+strconv.Itoa(t.serverPort))
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		logger.Error(err)
		t.atomic = 0
		return err
	}
	err = t.shake(conn, proto.CmdType, proto.InitiativeTransportUsage, lib.NextPosUid(), 0)
	if err != nil {
		conn.Close()
		t.atomic = 0
		return err
	}
	go t.handleCmd(conn)
	// TODO 连接池好像真的挺烦的，。。。
	// for i := 0; i < t.minConns; i++ {
	// 	go t.createConn()
	// }
	return nil
}

func (t *transport) checkReConn() {
	idles := atomic.LoadInt32(&(t.idleConns))
	if int(idles) < t.minConns {
		go t.createConn()
	}
}

func (t *transport) createConn() {
	addr, _ := net.ResolveTCPAddr("tcp", t.serverIp+":"+strconv.Itoa(t.serverPort))
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return
	}
	id := lib.NextPosUid()
	err = t.shake(conn, proto.TransportType, proto.InitiativeTransportUsage, id, 0)
	if err != nil {
		conn.Close()
		return
	}
	c := newWrappedConn(conn, id, true)
	c.run(t, true)
}

func (t *transport) monitor() {
	tk := time.NewTicker(time.Duration(time.Minute))
	for {
		<-tk.C
		logger.Debug("定时探测...")
		go t.createCmdAndConn()
	}
}

func (t *transport) handleCmd(conn net.Conn) {
	defer func() {
		conn.Close()
		t.atomic = 0
	}()
	for {
		cmd := proto.CmdProto{}
		err := cmd.Recv(conn)
		if err != nil {
			logger.Error(err)
			break
		}
		go func(c proto.CmdProto) {
			logger.Infof("获取到一个请求，用途=%d, reqID=%d", cmd.Usage, cmd.ReqID)
			switch cmd.Usage {
			case proto.TransportReqUsage:
				var conn net.Conn
				var err error
				var id int64
				addr, _ := net.ResolveTCPAddr("tcp", t.serverIp+":"+strconv.Itoa(t.serverPort))
				for i := 0; i < 3; i++ {
					if i == 2 {
						logger.Error("创建回应链接连续两次失败，不再尝试")
						return
					}
					conn, err = net.DialTCP("tcp", nil, addr)
					if err != nil {
						logger.Errorf("创建回应临时链接失败%v", err)
						continue
					}
					id = lib.NextPosUid()
					err = t.shake(conn, proto.TransportType, proto.TransportRspUsage,
						id, cmd.ReqID)
					if err != nil {
						continue
					}
					break
				}
				c := newWrappedConn(conn, id, true)
				c.run(t, false)
			}
		}(cmd)
	}

}

type wrappedConn1 struct {
	net.Conn
	duration bool
	id       int64
	atomic   int32
}

type wrappedConn struct {
	lib.WrapConnStru
}

func (w wrappedConn1) ID() int64 {
	return w.id
}

func (w *wrappedConn) Shutdown() error {
	return w.Shutdown()
}

func newWrappedConn(conn net.Conn, id int64, duration bool) wrappedConn {
	return wrappedConn{WrapConnStru: lib.NewWrapConn(conn, id)}
	// return wrappedConn{conn, duration, Ui, 0}
}

func (c wrappedConn) run(t *transport, reConn bool) {
	// c := wrappedConn{}
	addr, _ := net.ResolveTCPAddr("tcp", t.targetHost+":"+strconv.Itoa(t.targetPort))
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		logger.Errorf("建立与%s:%d的链接失败", t.targetHost, t.targetPort)
		c.Close()
		return
	}
	// shake := proto.ShakeProto{}
	// err = shake.Recv(c.Conn)
	// if err != nil {
	// 	logger.Error(err)
	// 	return
	// }
	wg := &sync.WaitGroup{}
	wc := lib.NewWrapConn(conn, lib.NextPosUid())
	// logger.Infof("%v id =%d", wc.Conn, wc.ID())
	wg.Add(2)
	go lib.Pipe2(wg, wc, c.WrapConnStru, func() {
		wc.Close()
		c.Close()
	})
	go lib.Pipe2(wg, c.WrapConnStru, wc, func() {
		wc.Close()
		c.Close()
	})
	if reConn {
		t.checkReConn()
	}
	wg.Wait()
	// wc.ShutDown()
	// c.Shutdown()
}
