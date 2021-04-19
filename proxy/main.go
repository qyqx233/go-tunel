package main

import (
	"flag"
	"net"

	"github.com/qyqx233/go-tunel/lib"
	"github.com/qyqx233/go-tunel/server"
)

var logger = lib.GetLogger()

func main() {
	var configPath string
	var isDump bool
	flag.StringVar(&configPath, "c", "proxy.toml", "config")
	flag.BoolVar(&isDump, "d", false, "is dump")
	flag.Parse()
	parseConfig(configPath)
	for _, transport := range config.Transport {
		svr := server.NewTransportServer(transport.Addr, transport.TargetHost, transport.TargetPort)
		server.With(func(transport *server.TransportServerStru) {
			transport.Handle = func(c net.Conn, buf []byte) {

			}
		})
		logger.Infof("在端口%s启动转发到%s:%d的服务", transport.Addr,
			transport.TargetHost, transport.TargetPort)
		err := svr.Start()
		if err != nil {
			panic(err)
		}
	}
	select {}
}
