package outer

import (
	"context"
	"sync"

	"github.com/qyqx233/go-tunel/lib"
	"github.com/qyqx233/go-tunel/lib/proto"
	"github.com/rs/zerolog/log"
)

const (
	RegState int = iota
)

type reqConnChanStru struct {
	reqID int64
	ch    *chan lib.WrapConnStru
}

type transportImpl struct {
	proxyStarted bool // 是否监听转发端口
	Dump         bool
	Export       bool
	AddIp        bool
	atomic       int32
	connNum      int32
	ID           uint64
	IP           string
	TargetHost   string
	TargetPort   int
	LocalPort    int
	TcpOrUdp     string
	Name         []byte
	SymKey       []byte
	MinConnNum   int
	MaxConnNum   int
	AllowIps     []string
	State        proto.ShakeStateEnum // 通道状态
	cmdConn      lib.WrapConnStru
	connCh       chan lib.WrapConnStru // 缓存的传输通道
	newCh        chan reqConnChanStru  // 用来监听是否需要创建临时通道
	asyncMap     sync.Map
	// cmdCh   chan struct{}
}

// 总是最多被一个coroutine调用
func (t *transportImpl) shutdown() {
	t.State = proto.ShutdownState
	for {
		select {
		case <-t.newCh:
		case ch := <-t.connCh:
			ch.ShutDown()
		default:
			return
		}
	}
}

// 总是最多被一个coroutine调用
func (t *transportImpl) restart(ctx context.Context, conn lib.WrapConnStru) {
	if t.State == proto.RegState {
		t.connCh = make(chan lib.WrapConnStru, t.MinConnNum)
		t.newCh = make(chan reqConnChanStru, 10)
	}
	t.cmdConn = conn
	go t.monitor(ctx)
	t.State = proto.ShakeState
}

func (t *transportImpl) monitor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ch := <-t.newCh:
			go func() {
				defer func() {
					if err := recover(); err != nil {
						log.Error().Msgf("%v", err)
					}
				}()
				rid := ch.reqID
				log.Info().Msgf("请求与服务器%s:%d的临时通道, reqID=%d", t.IP, t.TargetPort, rid)
				cmd := proto.CmdProto{Usage: proto.TransportReqUsage, ReqID: rid}
				err := cmd.Send(t.cmdConn.Conn)
				if err != nil {
					log.Error().Err(err).Msg("error")
					return
				}
				t.asyncMap.Store(rid, ch)
			}()
		}
	}
}

type transportList []*transportImpl

var initCap = 64

func (l transportList) search(h *transportImpl) (bool, int) {
	d := 0
	begin := 0
	end := len(l) - 1
	var mid int
	for begin <= end {
		mid = (begin + end) / 2
		// println(begin, end, mid)
		v := l[mid]
		if v.TargetHost == h.TargetHost && v.TargetPort == h.TargetPort {
			return true, mid
		} else if h.TargetHost < v.TargetHost || (h.TargetHost == v.TargetHost && h.TargetPort < v.TargetPort) {
			end = mid - 1
			d = 0
		} else {
			begin = mid + 1
			d = 1
		}
	}
	return false, mid + d
}

func (m *TransportMng) add(h *transportImpl) transportList {
	l := m.tl
	leng := len(l)
	if leng == 0 {
		l = append(l, h)
		m.tl = l
		return l
	}
	has, pos := l.search(h)
	if !has {
		l = append(l, &transportImpl{})
		for i := leng; i > pos; i-- {
			l[i] = l[i-1]
		}
		l[pos] = h
	}
	m.tl = l
	return l
}

func (m *TransportMng) remove(h *transportImpl) transportList {
	l := m.tl
	leng := len(l)
	if leng == 1 {
		l = make([]*transportImpl, 0, initCap)
		m.tl = l
		return l
	}
	_, pos := l.search(h)
	ll := make([]*transportImpl, leng-1, initCap)
	if pos > 0 {
		copy(ll[:pos], l[:pos])
	}
	copy(ll[pos:], l[pos+1:])
	l = ll
	m.tl = l
	return l
}

type TransportMng struct {
	rwl *sync.RWMutex
	tl  transportList
}

var tl transportList

var transportMng TransportMng

func init() {
	transportMng.tl = make([]*transportImpl, 0, initCap)
	transportMng.rwl = &sync.RWMutex{}
}
