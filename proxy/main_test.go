package main

import (
	"fmt"
	"net"
	"testing"
	"unsafe"
)

func foo(a []interface{}) {
	// fmt.Println(a)
	fmt.Println(fmt.Sprintf(a[0].(string), a[1:]...))
}

func IsLittleEndian() bool {
	var i uint = 0x01020304
	u := unsafe.Pointer(&i)
	pb := (*byte)(u)
	b := *pb
	return (b == 0x04)
}

type LogConn interface {
	net.Conn
	ID() uint64
}

func haha(c LogConn) {
	fmt.Println(c.ID())
}

type LogConnStru struct {
	net.Conn
	id uint64
}

func (c LogConnStru) ID() uint64 {
	return c.id
}

func Test_ss(t *testing.T) {
	c := wrapConn{id: 1100}
	c.log("11%d", 100)
	cc := LogConnStru{}
	haha(cc)
}
