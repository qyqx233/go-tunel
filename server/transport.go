package server

import (
	"gotunel/lib"
	"net"
	"strconv"
	"sync"

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

type TransportServerStru struct {
	conn net.Conn
	addr string
	host string
	port int
}

func NewTransportServer(addr string, host string, port int) TransportServerStru {
	t := TransportServerStru{nil, addr, host, port}
	return t
}

func (t TransportServerStru) Transport(conn net.Conn) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	conn1, err := NewTcpConn(t.host, t.port)
	if err != nil {
		logger.Error(err)
		conn.Close()
		return
	}
	wc := lib.NewWrapConn(conn, 0)
	wc1 := lib.NewWrapConn(conn1, 0)
	go lib.Pipe2(wg, wc, wc1, func() {
		conn.Close()
		conn1.Close()
	})
	lib.Pipe2(wg, wc1, wc, func() {
		conn.Close()
		conn1.Close()
	})
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
			go t.Transport(conn)
		}
	}()
	return nil
}
