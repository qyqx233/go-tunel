package proto

import (
	"net"
	"unsafe"

	"github.com/rs/zerolog/log"
)

type zz struct {
	Version byte
}

var Magic uint16 = 65424

type FileProto struct {
	Version int8
	Size    int8
	Name    [256]byte
	zz
}

type RawDataProto struct {
	Size int
	zz
}

type CmdProto struct {
	zz
	Type      int8
	Usage     ShakeProtoUsageEnum
	Code      ShakeErrEnum
	ReqID     int64
	CorrReqID int64
}

type ShakeProto struct {
	zz
	TcpOrUdp byte
	Type     TransportTypeEnum // 1-命令通道,2-传输通道
	Usage    ShakeProtoUsageEnum
	Code     ShakeErrEnum
	Magic    uint16
	Port     uint16
	Name     [16]byte
	SymKey   [16]byte
	// Error     [64]byte
	Host      [32]byte
	ReqID     int64
	CorrReqId int64
}

type ShakeStateEnum int

const (
	RegState      ShakeStateEnum = iota // 初始化状态
	ShakeState                          //握手成功的状态
	ShutdownState                       // 命令通道关闭的状态
)

type ShakeProtoUsageEnum int8

const (
	TransportReqUsage ShakeProtoUsageEnum = iota
	TransportRspUsage
	BeatUsage
	InitiativeTransportUsage
)

type TransportTypeEnum int8

const (
	CmdType TransportTypeEnum = iota
	TransportType
)

type ShakeErrEnum int8

const (
	OkCode ShakeErrEnum = iota
	HostNotRegisterCode
	KeyErrorCode
	TooManyConns
	MagicErrorCode
)

const (
	KeepAliveTcpOption int = iota
	AlivePeriodOption
	NoDelayOption
)

const (
	False int = iota
	True
)

func (cmd zz) send(conn net.Conn, p unsafe.Pointer, n int) error {
	slice := Slice{Addr: p, Cap: n, Len: n}
	bs := *(*[]byte)(unsafe.Pointer(&slice))
	// fmt.Println("开始发送", bs, len(bs), cap(bs))
	m, left, succ := 0, n, 0
	var err error
	for left > 0 {
		m, err = conn.Write(bs[succ:])
		if err != nil {
			return err
		}
		left -= m
		succ += m
	}
	return err
}

func (cmd zz) recv(conn net.Conn, p unsafe.Pointer, n int) error {
	slice := Slice{Addr: p, Cap: n, Len: n}
	bs := *(*[]byte)(unsafe.Pointer(&slice))
	var err error
	m, left, succ := 0, n, 0
	for left > 0 {
		m, err = conn.Read(bs[succ:])
		if err != nil {
			return err
		}
		left -= m
		succ += m
	}
	// fmt.Println("开始接收", bs, len(bs), cap(bs))
	return err
}

func (c *CmdProto) Send(conn net.Conn) error {
	return c.zz.send(conn, unsafe.Pointer(c), int(unsafe.Sizeof(*c)))
}

func (c *CmdProto) Recv(conn net.Conn) error {
	return c.zz.recv(conn, unsafe.Pointer(c), int(unsafe.Sizeof(*c)))
}

func (c *ShakeProto) Send(conn net.Conn) error {
	log.Debug().Uint16("magic", c.Magic).Msg("print magic num")
	return c.zz.send(conn, unsafe.Pointer(c), int(unsafe.Sizeof(*c)))
}

func (c *ShakeProto) Recv(conn net.Conn) error {
	return c.zz.recv(conn, unsafe.Pointer(c), int(unsafe.Sizeof(*c)))
}

type Slice struct {
	Addr unsafe.Pointer
	Len  int
	Cap  int
}

type String struct {
	Data unsafe.Pointer
	Len  int
}
