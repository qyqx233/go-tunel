package main

import (
	"flag"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	// "github.com/qyqx233/go-tunel/proxy/cmd"
	"github.com/qyqx233/go-tunel/server"
	"github.com/qyqx233/gtool/lib/convert"
	"github.com/rs/zerolog/log"
)

type writer struct {
	w io.WriteCloser
}

func newWriter(name string) *writer {
	fd, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	return &writer{
		w: fd,
	}
}

func (w *writer) Write(buf []byte) (int, error) {
	w.w.Write(convert.Int2Bytes(len(buf)))
	return w.w.Write(buf)
}

func (w *writer) Close() error {
	return w.w.(*os.File).Close()
}

type DumpServer struct {
	file string
	fd   *os.File
}

var ds *DumpServer

func (ds *DumpServer) content(w http.ResponseWriter, req *http.Request) {
	values := req.URL.Query()
	idStr := values.Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Error().Msgf("id = %s", idStr)
		w.Write([]byte("oh no"))
		return
	}
	ds.fd.Seek(0, io.SeekStart)
	length := make([]byte, 8)
	for i := 0; i < id; i++ {
		ds.fd.Read(length)
		n := convert.Bytes2Uint64(length)
		buf := make([]byte, n)
		ds.fd.Read(buf)
	}
	ds.fd.Read(length)
	n := convert.Bytes2Uint64(length)
	buf := make([]byte, n)
	ds.fd.Read(buf)
	w.Write(buf)
}

func newDumpServer(file string, port int) (err error) {
	ds = &DumpServer{file: file}
	ds.fd, err = os.Open(file)
	if err != nil {
		return err
	}
	http.HandleFunc("/content", ds.content)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
	return
}

func printMsg(file string) {
	fd, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	length := make([]byte, 8)
	for {
		var n int
		n, err = fd.Read(length)
		if err != nil || n == 0 {
			if err == io.EOF {
				break
			}
			log.Error().Err(err).Msg("error")
			break
		}
		data := make([]byte, n)
		fd.Read(data)
		println(convert.Bytes2String(data))
	}

}

func parseFlag(cmd, configPath, file *string, port *int, isDump *bool) {
	proxyCmd := flag.NewFlagSet("proxy", flag.ExitOnError)
	proxyCmd.StringVar(configPath, "c", "proxy.yaml", "config file")
	proxyCmd.IntVar(port, "p", 9088, "port")
	proxyCmd.StringVar(file, "f", "dump.bin", "config")
	proxyCmd.BoolVar(isDump, "d", false, "is dump")
	dumpCmd := flag.NewFlagSet("dump", flag.ExitOnError)
	dumpCmd.StringVar(file, "f", "dump.bin", "config")
	// flag.Parse()
	switch os.Args[1] {
	case "proxy":
		proxyCmd.Parse(os.Args[2:])
		*cmd = "proxy"
	case "dump":
		dumpCmd.Parse(os.Args[2:])
		*cmd = "dump"
	default:
		panic("no such cmd: " + "`" + os.Args[1] + "`")
	}
}

func main() {
	var configPath, file, cmd string
	var isDump bool
	var port int
	var w io.WriteCloser
	parseFlag(&cmd, &configPath, &file, &port, &isDump)
	parseConfig(configPath)
	// cmd.Execute()
	if cmd == "dump" {
		printMsg(file)
		return
	}

	if isDump {
		log.Info().Str("file", file).Msg("dump to file")
		w = newWriter(file)
	}
	defer w.Close()
	for _, transport := range config.Transport {
		addr := transport.Addr
		var host string
		var ports string
		var beginPort, endPort int
		tmpArr := strings.Split(addr, ":")
		host, ports = tmpArr[0], tmpArr[1]
		if strings.Contains(ports, "-") {
			tmpArr = strings.Split(ports, "-")
			beginPort, _ = strconv.Atoi(tmpArr[0])
			endPort, _ = strconv.Atoi(tmpArr[1])
		} else {
			beginPort, _ = strconv.Atoi(ports)
			endPort = beginPort
		}
		total := endPort - beginPort + 1
		ch := make(chan struct{}, 1000)
		go func() {
			for i := 0; i < total; i++ {
				ch <- struct{}{}
			}
			close(ch)
		}()
		for port := beginPort; port <= endPort; port++ {
			go func(port int) {
				<-ch
				svr := server.NewTransportServer(host+":"+strconv.Itoa(port), transport.TargetHost, transport.TargetPort)
				svr.Handle = func(c net.Conn, buf []byte) {
					log.Info().Str("local", c.LocalAddr().String()).Str("remote", c.RemoteAddr().String()).Msg("forward")
					w.Write(buf)
				}
				log.Info().Msgf("start server at %d, transport to %s:%d", port, transport.TargetHost, transport.TargetPort)
				err := svr.Start()
				if err != nil {
					log.Error().Err(err)
				}
			}(port)
		}
	}
	newDumpServer(file, port)
}
