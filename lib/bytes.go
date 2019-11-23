package lib

import (
	"gotunel/lib/proto"
	"unsafe"
)

func String2Byte32(s string) [32]byte {
	bytes := [32]byte{}
	var p *proto.String
	p = (*proto.String)(unsafe.Pointer(&s))
	sl := *(*[]byte)(unsafe.Pointer(&proto.Slice{
		Addr: p.Data,
		Len:  p.Len,
		Cap:  p.Len,
	}))
	copy(bytes[:p.Len], sl)
	return bytes
}

func String2Byte16(s string) [16]byte {
	bytes := [16]byte{}
	var p *proto.String
	p = (*proto.String)(unsafe.Pointer(&s))
	sl := *(*[]byte)(unsafe.Pointer(&proto.Slice{
		Addr: p.Data,
		Len:  p.Len,
		Cap:  p.Len,
	}))
	copy(bytes[:p.Len], sl)
	return bytes
}

func Byte16ToBytes(arr [16]byte) []byte {
	bs := make([]byte, 0, 16)
	for i := 0; i < 16; i++ {
		if arr[i] != 0 {
			bs = append(bs, arr[i])
		} else {
			break
		}
	}
	return bs
}

func Byte32ToBytes(arr [32]byte) []byte {
	bs := make([]byte, 0, 32)
	for i := 0; i < 32; i++ {
		if arr[i] != 0 {
			bs = append(bs, arr[i])
		} else {
			break
		}
	}
	return bs
}

func Uint642String(u uint64) string {
	hexChars := "0123456789abcdef"
	// fmt.Println(unsafe.Sizeof(i))
	var bb []byte
	flag := false
	for i := uint64(0); i < uint64(unsafe.Sizeof(u)); i++ {
		x := u >> ((7 - i) * 8) & 255
		if !flag && x != 0 {
			flag = true
		}
		if !flag && x == 0 {
			continue
		}
		bb = append(bb, hexChars[x>>4])
		bb = append(bb, hexChars[x&15])
	}
	if len(bb) == 0 {
		bb = append(bb, '0')
	}
	return string(bb)
}
