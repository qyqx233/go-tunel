package proto

import (
	"fmt"
	"net"
	"testing"
)

func foo(inf interface{}) {

}

type aa interface {
	net.Conn
}

type aas struct {
	net.Conn
}

func (a aas) xx() {

}

func xxx(a aa) {
}
func Test1(t *testing.T) {
	p := CmdProto{}
	fmt.Printf("%p %p\n", &p, &p.zz)
	foo(&p)
	fmt.Println([]byte("12341234"))
	fmt.Println(HostNotRegisterCode)
	var a aas
	xxx(a)
	a.Close()
}
