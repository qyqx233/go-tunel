package outer

import (
	"net"
	"sync"

	"github.com/qyqx233/go-tunel/lib"
	"github.com/qyqx233/go-tunel/lib/file"
)

var dw *file.DumpWriter

func init() {
	dw = file.NewDumpWriter("dump.bin")
}

func writeLog(c net.Conn, data []byte) {
	dw.Write(data)
}

func pipeSocketWithLog(wg *sync.WaitGroup, wc, wt lib.WrapConnStru) {
	go lib.Pipe3(wg, wc, wt, func() {
		wc.ShutDown()
		wt.ShutDown()
	}, writeLog)
	go lib.Pipe3(wg, wt, wc, func() {
		wt.ShutDown()
		wc.ShutDown()
	}, writeLog)
}

func pipeSocket(wg *sync.WaitGroup, wc, wt lib.WrapConnStru) {
	go lib.Pipe2(wg, wc, wt, func() {
		wc.ShutDown()
		wt.ShutDown()
	})
	go lib.Pipe2(wg, wt, wc, func() {
		wt.ShutDown()
		wc.ShutDown()
	})
}
