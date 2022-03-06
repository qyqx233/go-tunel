package rest

import (
	"bytes"
	"testing"
)

func TestDecoder(t *testing.T) {
	var ts = TransportPdb{true, false}
	t.Log(ts.encode())
	var ts1 TransportPdb
	ts1.decode(ts.encode())
	t.Log(ts1)
}

func TestDB(t *testing.T) {
	InitDB()
	key := []byte("hosts:")
	var hosts [][]byte
	oldValue, _ := PebbleGetBytes(key)
	if len(oldValue) > 0 {
		hosts = bytes.Split(oldValue, []byte(":"))
	}
	t.Log(hosts)
}

func TestBytes(t *testing.T) {
	var v = []byte("")
	t.Log(bytes.Split(v, []byte(":")))
}
