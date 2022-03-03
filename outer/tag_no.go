// +build !WITH_LOG

package outer

import (
	"net"
	"sync"

	"github.com/qyqx233/go-tunel/lib"
)

// var dw *file.DumpWriter

func WriteLog(data []byte) {
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