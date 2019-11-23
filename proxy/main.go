package main

import (
	"fmt"
	"godemo/tunel/lib"
	"net"
	"strconv"
	"sync"
	"time"
)

type connSvrStru struct {
	localPort int
	host      string
	port      int
}

var uid uint64

type wrapConn struct {
	net.Conn
	id int64
}

func (c wrapConn) ID() int64 {
	return c.id
}

func (c wrapConn) log(vars ...interface{}) {
	f := vars[0].(string)
	f = strconv.Itoa(int(c.id)) + ": " + f
	logger.Infof(f, vars[1:])
}

var logger = lib.GetLogger()

func (c connSvrStru) handleConnConn(conn net.Conn) {
	defer conn.Close()
	var transport net.Conn

	addr, _ := net.ResolveTCPAddr("tcp", c.host+":"+strconv.Itoa(c.port))
	transport, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return
	}
	wc, wt := lib.NewWrapConn(conn, 0), lib.NewWrapConn(transport, 0)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go lib.Pipe2(wg, wc, wt, func() {
		conn.Close()
		transport.Close()
	})
	go lib.Pipe2(wg, wt, wc, func() {
		transport.Close()
		conn.Close()
	})
	wg.Wait()
}

func (c connSvrStru) start() error {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(c.localPort))
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			go c.handleConnConn(conn)
		}
	}()
	return nil
}

func main() {
	// c := connSvrStru{localPort: 5000, host: "cocoing.tk", port: 4403}
	// c.start()
	// select {}
	fmt.Println("begin")
	t := time.NewTimer(time.Duration(time.Second * 20))
	for {
		<-t.C
		t.Reset(time.Duration(time.Second * 1))
	}
}
