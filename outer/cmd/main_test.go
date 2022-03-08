package main

import (
	"bytes"
	"sort"
	"testing"

	"github.com/qyqx233/go-tunel/outer/rest"
)

func TestDB(t *testing.T) {
	key := []byte("hosts:")
	var hosts [][]byte
	oldValue, _ := rest.PebbleGetBytes(key)
	if len(oldValue) > 0 {
		hosts = bytes.Split(oldValue, []byte(":"))
	}
	var index = sort.Search(len(hosts), func(i int) bool {
		return bytes.Compare(hosts[i], []byte("g.com")) >= 0
	})
	t.Log(index)
	t.Log(hosts)
}

func Test_BinarySearch(t *testing.T) {
	var s = []int{1, 2, 3, 10}
	t.Log(rest.SearchSlice(s, 1))
	t.Log(rest.SearchSlice(s, 2))
	t.Log(rest.SearchSlice(s, 3))
	s = []int{1}
	t.Log(rest.SearchSlice(s, 1))
	t.Log(rest.SearchSlice(s, 0))

	// a := []byte("a")
	// sl := [][]byte{a}
	// rest.SearchComparableSlice[[]byte]([]rest.Bytes(sl), a)
}

type Int int

func add(i Int) Int {
	return i + 1
}

func sum(is []Int) Int {
	return is[0]
}

func Test_type(t *testing.T) {
	var i = 19009
	add(Int(i))
	sum([]Int{1})
	var a = []byte("a")
	var b = []byte("b")
	var bs = [][]byte{a}
	for _, x := range [][]byte{a, b} {
		index := rest.SearchSliceFunc(len(bs), func(i int) int {
			return bytes.Compare(bs[i], x)
		})
		t.Log(index)
	}
	// sum([]int{1})
}

type adder interface {
	add()
}

type minuser interface {
	minus()
}

func Haha[T adder](t T) {

}
