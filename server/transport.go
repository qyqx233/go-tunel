package server

import (
	"net"
	"strconv"
	"sync"

	"github.com/qyqx233/go-tunel/lib"
	"github.com/rs/zerolog/log"
)

func NewTcpConn(host string, port int) (conn net.Conn, err error) {
	var addr *net.TCPAddr
	addr, err = net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		return
	}
	conn, err = net.DialTCP("tcp", nil, addr)
	return
}

type HandleFunc = func(net.Conn, []byte)

type TransportServerImpl struct {
	conn   net.Conn
	addr   string
	host   string
	port   int
	Handle HandleFunc
}

func With(f func(transport *TransportServerImpl)) {
}

func NewTransportServer(addr string, host string, port int) *TransportServerImpl {
	t := TransportServerImpl{nil, addr, host, port, nil}
	return &t
}

func NewTransportDebugServer(addr string, host string, port int) TransportServerImpl {
	t := TransportServerImpl{nil, addr, host, port, nil}
	return t
}

func (t *TransportServerImpl) TransportExt(conn net.Conn) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	conn1, err := NewTcpConn(t.host, t.port)
	if err != nil {
		log.Error().Err(err).Msg("error")
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
	log.Info().Msg("链接断开")
	wg.Wait()
}

func (t *TransportServerImpl) Transport(conn net.Conn) {
	wg := &sync.WaitGroup{}
	conn1, err := NewTcpConn(t.host, t.port)
	if err != nil {
		log.Error().Err(err).Msg("error")
		conn.Close()
		return
	}
	wg.Add(2)
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
	log.Info().Msg("链接断开")
	wg.Wait()
}

func (t TransportServerImpl) Shutdown() {
	t.conn.Close()
}

func (t TransportServerImpl) Start() error {
	listener, err := net.Listen("tcp", t.addr)
	if err != nil {
		log.Error().Err(err).Msg("error")
		return err
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			log.Info().Msg("客户端地址:" + conn.RemoteAddr().String())
			go t.Transport(conn)
		}
	}()
	return nil
}
