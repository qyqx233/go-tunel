package outer

import (
	"bytes"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"unsafe"

	"github.com/qyqx233/go-tunel/lib/proto"
)

type _ss struct {
	_ [64]byte
}

func TestList(t *testing.T) {
	s := make([]int, 0, 10)
	s = append(s, []int{0, 1, 2, 3, 4}...)
	copy(s[3:6], s[2:5])
	// var hl hostList = make([]*transportStru, 0, 10)
	h := transportImpl{IP: "1.1.1.1", TargetPort: 6002}
	fmt.Printf("%p\n", &h)
	h = transportImpl{IP: "1.1.1.1", TargetPort: 6003}
	fmt.Printf("%p\n", &h)
	h1 := &transportImpl{IP: "1.1.1.1", TargetPort: 6002}
	fmt.Printf("%p\n", &h1)
	h1 = &transportImpl{IP: "1.1.1.1", TargetPort: 6003}
	fmt.Printf("%p\n", &h1)

	// hl = hl.add(h)
	// t.Log(hl)
	// h = &transportStru{Host: "1.1.1.1", Port: 6003}
	// // t.Log(&h1)
	// hl = hl.add(h)
	// t.Log(hl)
	// t.Log(unsafe.Sizeof(hl))
	// _, pos := hl.search(h)
	// t.Log(pos)
	// t.Log(hl[pos].Port)
}

func TestBytes(t *testing.T) {
	cmd := proto.ShakeProto{}
	t.Log(unsafe.Sizeof(cmd))
	slice := proto.Slice{Addr: unsafe.Pointer(&cmd), Cap: int(unsafe.Sizeof(cmd)),
		Len: int(unsafe.Sizeof(cmd))}
	data := *(*[]byte)(unsafe.Pointer(&slice))
	fmt.Println(data)
}

func TestMap(t *testing.T) {
	mp := sync.Map{}
	mp.Store("a", 100)
	fmt.Println(&mp)
	a, b := 1, 200
	var inf interface{} = (b - a) / 3
	fmt.Println(reflect.TypeOf(inf))
}

func Test_srvMng(t *testing.T) {
	mng := newproxySvrMng(3000, 4000)
	t.Log(mng.findAvailablePort())
	t.Log(mng.findAvailablePort())
	reqID = 100
	t.Log(mng.findAvailablePort())
}

func Test_map1(t *testing.T) {
	var mp sync.Map
	mp.Store(1, 1)
	fmt.Println(mp.Load(1))
	fmt.Println(mp.Load(11))
}

type sss struct {
	a int
}

func Test_chan(t *testing.T) {

	ch := make(chan *sss, 2)
	var sl = []sss{{1}, {2}}
	go func() {
		for i := 0; i < 2; i++ {
			fmt.Printf("%p\n", &sl[i])
			ch <- &sl[i]
		}
	}()
	// for {
	// 	select {
	// 	case s := <-ch:
	// 		fmt.Printf("%p\n", &s)
	// 	}
	// }
	for x := range ch {
		fmt.Printf("%p %d\n", &x, x)
		println(x)
	}
}

func TestInt(t *testing.T) {
	var y = 1222_1233_3341
	fmt.Println(y)
	var maxInt64 int64 = 2<<62 - 1 // 63个1
	fmt.Println(^maxInt64)
	var maxUint64 uint64 = 2<<63 - 1
	var maxUint63 uint64 = 2<<62 - 1
	fmt.Println(maxInt64, maxUint64, maxUint64&maxUint63)
	var inf interface{} = maxInt64 & maxInt64
	fmt.Println(reflect.TypeOf(inf))
	var x = *(*int64)(unsafe.Pointer(&maxUint64))
	// var i64 int64 = 1
	fmt.Println(x, int(maxUint63&1)*-1)
}

func TestConfig(t *testing.T) {
	parseConfig("outer.toml")
	t.Log(config.Transport[0].IP, config.Transport[0].TargetPort, config.Transport[0].Symkey)
	t.Log(config.CmdServer.Port)
	t.Log(config.ProxyServer.MinPort)
	t.Log(config.HttpServer.Port)
	bytes.Equal([]byte(""), []byte(""))

}
