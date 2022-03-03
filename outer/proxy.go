package outer

import (
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qyqx233/go-tunel/lib"
	"github.com/rs/zerolog/log"
)

type proxySvrMngStru struct {
	m                sync.RWMutex
	minPort, maxPort int
	pushPos, popPos  int
	ports            []int32
	proxySvrs        map[int]proxySvrStru
}

var proxySvrMng *proxySvrMngStru

func newproxySvrMng(minPort, maxPort int) *proxySvrMngStru {
	return &proxySvrMngStru{
		minPort:   minPort,
		maxPort:   maxPort,
		ports:     make([]int32, 0, maxPort-minPort+1),
		proxySvrs: make(map[int]proxySvrStru),
	}
}

func (m *proxySvrMngStru) findAvailablePort() int {
	secs := ((m.maxPort + 1 - m.minPort) / 64)
	sec := int(reqID) % secs
	ports := m.ports[64*sec : 64*(sec+1)]
	for i := 0; i < 64; i++ {
		if !atomic.CompareAndSwapInt32(&ports[i], 0, 1) {
			continue
		}
		return sec*64 + i + m.minPort
	}
	return -1
}

func (m *proxySvrMngStru) setPortavailable() int {
	secs := ((m.maxPort + 1 - m.minPort) / 64)
	sec := int(reqID) % secs
	ports := m.ports[64*sec : 64*(sec+1)]
	for i := 0; i < 64; i++ {
		if !atomic.CompareAndSwapInt32(&ports[i], 0, 1) {
			continue
		}
		return sec*64 + i + m.minPort
	}
	return -1
}

func (m *proxySvrMngStru) newServer(t *transportImpl) error {
	var port = t.LocalPort
	if port == 0 {
		port = m.findAvailablePort()
		t.LocalPort = port
	}
	svr := proxySvrStru{localPort: port, t: t}
	log.Error().Msgf("在端口%d启动服务转发至%s:%s:%d", port, t.IP, t.TargetHost, t.TargetPort)
	err := svr.start()
	if err != nil {
		return err
	}
	m.m.Lock()
	m.proxySvrs[port] = svr
	m.m.Unlock()
	return nil
}

type proxySvrStru struct {
	localPort int
	t         *transportImpl
}

func (c *proxySvrStru) handleConnConn(conn net.Conn) {
	var wt lib.WrapConnStru
	var ch chan lib.WrapConnStru
	var t *time.Timer
	var wc = lib.NewWrapConn(conn, lib.NextUid())
	select {
	case wt = <-c.t.connCh:
	default:
		log.Info().Msg("需要获取临时通道")
		t = time.NewTimer(time.Duration(int64(time.Second) * config.Global.WaitNewConnSeconds))
		ch = make(chan lib.WrapConnStru)
		c.t.newCh <- reqConnChanStru{wc.ID(), &ch}
		// wt = <-ch
		select {
		case wt = <-ch:
			t.Stop()
		case <-t.C:
			log.Error().Msg("获取临时通道超时")
			conn.Close()
			return
		}
	}
	wg := &sync.WaitGroup{}
	wg.Add(2)
	pipeSocket(wg, wc, wt)
	wg.Wait()
}

func (c proxySvrStru) start() error {
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
