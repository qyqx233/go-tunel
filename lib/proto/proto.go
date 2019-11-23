package proto

import (
	"net"
	"unsafe"
)

type zz struct {
	Version byte
}

type ShakeProto struct {
	TcpOrUdp byte
	Type     int8 // 1-命令通道,2-传输通道
	Usage    int8
	Code     int8
	Name     [16]byte
	SymKey   [16]byte
	// Error     [64]byte
	Host      [32]byte
	Port      int
	ReqID     int64
	CorrReqId int64
	zz
}

const (
	RegState      int = iota // 初始化状态
	ShakeState               //握手成功的状态
	ShutdownState            // 命令通道关闭的状态
)

const (
	TransportReqUsage int8 = iota
	TransportRspUsage
	BeatUsage
	InitiativeTransportUsage
)

const (
	CmdType int8 = iota
	TransportType
)

const (
	OkCode int8 = iota
	HostNotRegisterCode
	KeyErrorCode
	TooManyConns
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

type CmdProto struct {
	zz
	Type      int8
	Usage     int8
	Code      int8
	ReqID     int64
	CorrReqID int64
}

func (cmd zz) send(conn net.Conn, p unsafe.Pointer, n int) error {
	slice := Slice{Addr: p, Cap: n, Len: n}
	bs := *(*[]byte)(unsafe.Pointer(&slice))
	// fmt.Println("开始发送", bs, len(bs), cap(bs))
	_, err := conn.Write(bs)
	return err
}

func (cmd zz) recv(conn net.Conn, p unsafe.Pointer, n int) error {
	slice := Slice{Addr: p, Cap: n, Len: n}
	bs := *(*[]byte)(unsafe.Pointer(&slice))
	_, err := conn.Read(bs)
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
