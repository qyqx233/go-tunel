package main

import (
	"flag"
	"gotunel/lib"
	"gotunel/server"
	"strconv"
)

var logger = lib.GetLogger()

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "proxy.toml", "config")
	flag.Parse()
	for _, transport := range config.Transport {
		svr := server.NewTransportServer(":"+strconv.Itoa(transport.ListenPort),
			transport.TargetHost, transport.TargetPort)
		logger.Infof("在端口%d启动转发到%s:%d的服务", transport.ListenPort,
			transport.TargetHost, transport.TargetPort)
		err := svr.Start()
		if err != nil {
			panic(err)
		}
	}
	select {}
}
