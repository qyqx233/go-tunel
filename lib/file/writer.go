package file

import (
	"bufio"
	"os"
	"time"
	"unsafe"

	"github.com/qyqx233/gtool/lib/convert"
)

type DumpWriter struct {
	w *bufio.Writer
	f *os.File
}

func NewDumpWriter(name string) *DumpWriter {
	fd, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	return &DumpWriter{
		w: bufio.NewWriter(fd),
	}
}

type MsgHeader struct {
	Len       int
	Timestamp int64
}

const headerLen = int(unsafe.Sizeof(MsgHeader{}) - unsafe.Sizeof(int(0)))

// func stru2Bytes(p uintptr, len int) []byte {
// 	sl := reflect.SliceHeader{}
// 	return
// }

func (w *DumpWriter) Write(buf []byte) (int, error) {
	var header = MsgHeader{}
	header.Len = len(buf) + headerLen
	header.Timestamp = time.Now().Unix()
	w.w.Write(convert.Int2Bytes(len(buf) + int(unsafe.Sizeof(int(0)))))
	w.w.Write(convert.Int642Bytes(header.Timestamp))
	return w.w.Write(buf)
}

func (w *DumpWriter) Close() error {
	w.w.Flush()
	return w.f.Close()
}
