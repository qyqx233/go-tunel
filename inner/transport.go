package inner

import (
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qyqx233/go-tunel/lib"
	"github.com/qyqx233/go-tunel/lib/proto"
	"github.com/rs/zerolog/log"
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
	atomic     int32
	shakeRetry int
}

func (t *transport) shake(conn net.Conn, transportType proto.TransportTypeEnum, usage proto.ShakeProtoUsageEnum, reqID int64, corrReqID int64) error {
	shake := proto.ShakeProto{
		Magic:     proto.Magic,
		Type:      transportType,
		Usage:     usage,
		Name:      t.name,
		SymKey:    t.symkey,
		Host:      lib.String2Byte32(t.targetHost),
		Port:      uint16(t.targetPort),
		ReqID:     reqID,
		CorrReqId: corrReqID,
	}
	err := shake.Send(conn)
	if err != nil {
		log.Error().Err(err).Msg("error")
		return err
	}
	if transportType == proto.CmdType {
		err = shake.Recv(conn)
		if err != nil {
			log.Error().Err(err).Msg("error")
			return err
		}
		if shake.Code != proto.OkCode {
			log.Info().Msgf("握手返回错误码%d", shake.Code)
			return FatalError{code: int8(shake.Code), msg: "握手错误"}
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
		log.Error().Err(err).Msg("error")
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

func (t *transport) monitor(wg *sync.WaitGroup) {
	tk := time.NewTicker(time.Duration(time.Minute))
	for {
		<-tk.C
		log.Debug().Msg("定时探测...")
		err := t.createCmdAndConn()
		if _, ok := err.(FatalError); ok {
			tk.Stop()
			wg.Done()
			break
		}
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
			log.Error().Err(err).Msg("error")
			break
		}
		go func(c proto.CmdProto) {
			log.Info().Msgf("获取到一个请求，用途=%d, reqID=%d", cmd.Usage, cmd.ReqID)
			switch cmd.Usage {
			case proto.TransportReqUsage:
				var conn net.Conn
				var err error
				var id int64
				addr, _ := net.ResolveTCPAddr("tcp", t.serverIp+":"+strconv.Itoa(t.serverPort))
				for i := 0; i < 3; i++ {
					if i == 2 {
						log.Info().Msgf("创建回应链接连续两次失败，不再尝试")
						return
					}
					conn, err = net.DialTCP("tcp", nil, addr)
					if err != nil {
						log.Info().Msgf("创建回应临时链接失败%v", err)
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

func newWrappedConn(conn net.Conn, id int64, duration bool) *wrappedConn {
	return &wrappedConn{WrapConnStru: lib.NewWrapConn(conn, id)}
	// return wrappedConn{conn, duration, Ui, 0}
}

func (c *wrappedConn) run(t *transport, reConn bool) {
	// c := wrappedConn{}
	var target = t.targetHost + ":" + strconv.Itoa(t.targetPort)
	addr, _ := net.ResolveTCPAddr("tcp", target)
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Info().Msgf("建立与%s:%d的链接失败", t.targetHost, t.targetPort)
		c.Close()
		return
	}
	log.Debug().Str("target", target).Msg("connection made")
	wg := &sync.WaitGroup{}
	wc := lib.NewWrapConn(conn, lib.NextPosUid())
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
}
