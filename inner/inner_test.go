package main

import (
	"net"
	"reflect"
	"testing"
)

func Test1(t *testing.T) {
	addr, _ := net.ResolveTCPAddr("tcp", ":3333")
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		logger.Error(err)
	}
	t.Log(reflect.TypeOf(conn))
	// conn.Re
	// if _, ok := conn.(io.WriterTo); ok {
	// 	logger.Debug("WriterTo")

	// }
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	// if _, ok := conn.(io.ReaderFrom); ok {
	// 	logger.Debug("ReaderFrom")
	// }
}
