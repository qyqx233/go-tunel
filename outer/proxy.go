package outer

import (
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qyqx233/go-tunel/lib"
	"github.com/qyqx233/go-tunel/outer/pub"
	"github.com/qyqx233/go-tunel/outer/rest"
	"github.com/rs/zerolog/log"
)

type proxySvrMngSt struct {
	m                sync.RWMutex
	minPort, maxPort int
	ports            []int32
	proxySvrs        map[int]*proxySvrSt
}

var proxySvrMng *proxySvrMngSt

func newproxySvrMng(minPort, maxPort int) *proxySvrMngSt {
	return &proxySvrMngSt{
		minPort:   minPort,
		maxPort:   maxPort,
		ports:     make([]int32, 0, maxPort-minPort+1),
		proxySvrs: make(map[int]*proxySvrSt),
	}
}

// 读取ch更新server状态（是否关闭etc...）
// 前提假设, `proxySvrs`被正常初始化
func (m *proxySvrMngSt) handle() {
	go func() {
		for {
			req := <-rest.EnableSvrChan
			if svr, ok := m.proxySvrs[int(req.Port)]; ok {
				if req.Enable {
					if !svr.t.proxyStarted {
						err := svr.start()
						if err == nil {
							svr.t.proxyStarted = true
						}
					}
				} else {
					if svr.t.proxyStarted {
						svr.stop()
						svr.t.proxyStarted = false
					}
				}
			}
		}
	}()

}

func (m *proxySvrMngSt) findAvailablePort() int {
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

func (m *proxySvrMngSt) setPortavailable() int {
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

func (m *proxySvrMngSt) newServer(t *transportImpl) error {
	var port = t.LocalPort
	if port == 0 {
		port = m.findAvailablePort()
		t.LocalPort = port
	}
	svr := &proxySvrSt{localPort: port, t: t}
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

func (m *proxySvrMngSt) closeServer(t *transportImpl) (err error) {
	err = m.proxySvrs[t.LocalPort].stop()
	return
}

type proxySvrSt struct {
	localPort int
	t         *transportImpl
	listener  net.Listener
}

func (c *proxySvrSt) handleConnConn(conn net.Conn) {
	var wt lib.WrapConnStru
	var ch chan lib.WrapConnStru
	var t *time.Timer
	// 如果enable=False 表示当前转发通道并暂时关闭，此时直接关闭socket
	if transport, ok := pub.MemStor.Transports[c.t.LocalPort]; ok {
		log.Debug().Interface("transport", transport).Msg("print transport")
		if !transport.Enable {
			conn.Close()
			return
		}
	}

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
	if c.t.Dump {
		pipeSocketWithLog(wg, wc, wt)
	} else {
		pipeSocket(wg, wc, wt)
	}
	wg.Wait()
}

func (svr *proxySvrSt) start() error {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(svr.localPort))
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			go svr.handleConnConn(conn)
		}
	}()
	svr.listener = listener
	return nil
}

func (svr *proxySvrSt) stop() (err error) {
	svr.t.shutdown()
	err = svr.listener.Close()
	if err != nil {
		log.Error().Err(err).Int("port", svr.localPort).Msg("关闭转发端口失败")
		return
	}
	svr.t.proxyStarted = false
	log.Info().Int("port", svr.localPort).Msg("关闭转发端口成功")
	return
}
