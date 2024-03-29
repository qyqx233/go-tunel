package outer

import (
	"bytes"
	"context"
	"errors"
	"net"
	"strings"
	"sync/atomic"

	"github.com/qyqx233/go-tunel/lib"
	"github.com/qyqx233/go-tunel/lib/proto"
	"github.com/qyqx233/go-tunel/outer/pub"

	"github.com/rs/zerolog/log"
)

var reqID uint64

// 负责接收内网机器发过来的命令通道与传输通道
// 命令通道建立成功之后会分配一个端口用来转发请求至内网机器。如果命令通道断开，并不会关闭该端口
type cmdServer struct {
}

func getIp(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx >= 0 {
		return s[:idx]
	}
	return s
}

// 握手成功后会创建服务器
func (c *cmdServer) cmdLoop(wc lib.WrapConnStru, shake *proto.ShakeProto, t *transportImpl) {
	defer wc.Close()
	log.Info().Msg("开始回应命令通道")
	lib.SetTcpOptions(wc.Conn, proto.KeepAliveTcpOption, proto.True, proto.NoDelayOption, proto.True)
	// 收到成功响应内网客户端才认为通道建立成功
	shake.Code = proto.OkCode
	err := shake.Send(wc)
	if err != nil {
		log.Error().Msgf("响应命令通道握手信息失败 %v", err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if !t.proxyStarted {
		err = proxySvrMng.newServer(t)
		if err != nil {
			log.Error().Msgf("监听端口失败 %v", err) // 如果监听失败，则直接关闭链接，等待下一次重试
			return
		}
		t.proxyStarted = true
	}
	if t.AddIp {
		pub.MemStor.Ips[getIp(wc.RemoteAddr().String())] = struct{}{}
	}
	t.restart(ctx, wc) // 把conn传递给transportStru
	defer func() {
		t.shutdown()
	}()
	for {
		cmd := proto.CmdProto{} // 接收内网客户端发过来的心跳包
		err = cmd.Recv(wc)
		if err != nil {
			log.Error().Err(err).Msg("error")
			break
		}
		if cmd.Usage == proto.BeatUsage {
			log.Info().Msgf("接收到通道%d发过来的心跳", wc.ID())
			cmd.Code = proto.OkCode
			err = cmd.Send(wc)
			if err != nil {
				log.Info().Msgf("回应通道%d心跳失败", wc.ID())
				return
			}
		}
	}
}

func (c *cmdServer) shake(conn net.Conn, shake *proto.ShakeProto) (string, error) {
	addr := conn.RemoteAddr().String()
	hostPorts := strings.Split(addr, ":")
	host := hostPorts[0]
	err := shake.Recv(conn)
	return host, err
}

func (c *cmdServer) auth(conn net.Conn, host string, shake *proto.ShakeProto) (t *transportImpl) {
	targetHost := string(lib.Byte32ToBytes(shake.Host))
	log.Info().Msgf("服务器 地址%s:%s:%d", host, targetHost, shake.Port)
	if shake.Magic != proto.Magic {
		log.Info().Uint16("magic", shake.Magic).Msg("proto magic error")
		shake.Code = proto.MagicErrorCode
		return
	}
	v := transportImpl{IP: host,
		TargetPort: int(shake.Port),
		TargetHost: targetHost,
		SymKey:     lib.Byte16ToBytes(shake.SymKey),
		Name:       lib.Byte16ToBytes(shake.Name),
	}
	has, pos := transportMng.tl.search(&v)
	if has {
		t = transportMng.tl[pos]
		if !bytes.Equal(v.SymKey, t.SymKey) {
			log.Info().Msg("key不匹配")
			shake.Code = proto.KeyErrorCode
		}
	} else {
		log.Error().Str("host", v.TargetHost).Int("port", v.TargetPort).Msg("查询不到服务器")
		log.Error().Interface("transport List", transportMng.tl).Msg("通道列表")
		shake.Code = proto.HostNotRegisterCode
	}
	return
}

func (c *cmdServer) dispatch(conn net.Conn, shake *proto.ShakeProto, t *transportImpl) (bool, error) {
	var needClose = true
	if shake.Type == proto.TransportType { // 如果是传输通道
		needClose = false
		if shake.Usage == proto.InitiativeTransportUsage {
			select {
			case t.connCh <- lib.NewWrapConn(conn, shake.ReqID):
				atomic.AddInt32(&(t.connNum), 1)
			default:
				shake.Code = proto.TooManyConns
				return true, errors.New("too many conns")
			}
		} else if shake.Usage == proto.TransportRspUsage {
			log.Info().Msgf("获取到一个回应通道%s，reqId=%d, 原reqID=%d", conn.RemoteAddr().String(),
				shake.ReqID, shake.CorrReqId)
			lib.SetTcpOptions(conn, proto.KeepAliveTcpOption, proto.True,
				proto.NoDelayOption, proto.True)
			v, ok := t.asyncMap.Load(shake.CorrReqId)
			if ok {
				ch1 := v.(reqConnChanStru).ch
				*ch1 <- lib.NewWrapConn(conn, shake.ReqID)
				t.asyncMap.Delete(shake.CorrReqId)
			}
		}
	} else {
		if !atomic.CompareAndSwapInt32(&t.atomic, 0, 1) {
			log.Info().Msg("其他命令通道存活或处理中")
			return true, errors.New("其他命令通道存活或处理中")
		}
	}
	return needClose, nil
}

func (c *cmdServer) handleConn(conn net.Conn) {
	var t *transportImpl
	var shake proto.ShakeProto
	var needClose bool = true
	host, err := c.shake(conn, &shake)
	if err != nil {
		log.Error().Err(err).Msg("error")
		conn.Close()
		return
	}
	wc := lib.NewWrapConn(conn, shake.ReqID)
	t = c.auth(conn, host, &shake)
	if shake.Code != proto.OkCode {
		shake.Send(conn)
		wc.ShutDown()
		return
	}
	needClose, err = c.dispatch(conn, &shake, t)
	if err != nil {
		shake.Send(conn)
		wc.ShutDown()
		return
	}
	defer func() {
		if needClose {
			log.Info().Msg("命令通道关闭")
			wc.ShutDown()
		}
		if shake.Type == proto.CmdType {
			t.atomic = 0
		}
	}()
	if shake.Type == proto.CmdType {
		c.cmdLoop(wc, &shake, t)
	}
}

func newCmdServer(addr string) (*cmdServer, error) {
	cmdSvr := cmdServer{}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			go cmdSvr.handleConn(conn)
		}
	}()
	return &cmdServer{}, nil
}

func (cs *cmdServer) start() {
}
