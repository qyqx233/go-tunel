package lib

import (
	"testing"
)

func Test1(t *testing.T) {
	var s = "12341234"
	// var a = 1100
	// fmt.Printf("%p\n", &a)
	// p := (*proto.String)(unsafe.Pointer(&s))

	t.Log(String2Byte16(s))
	var bb [16]byte
	bb[0] = 'a'
	bb[1] = 'a'
	// fmt.Println(string(Byte162Bytes(bb)))
}
