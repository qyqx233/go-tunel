package main

import (
	"flag"
	"gotunel/lib"
	"gotunel/server"
)

var logger = lib.GetLogger()

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "proxy.toml", "config")
	flag.Parse()
	parseConfig(configPath)
	for _, transport := range config.Transport {
		svr := server.NewTransportServer(transport.Addr, transport.TargetHost, transport.TargetPort)
		logger.Infof("在端口%s启动转发到%s:%d的服务", transport.Addr,
			transport.TargetHost, transport.TargetPort)
		err := svr.Start()
		if err != nil {
			panic(err)
		}
	}
	select {}
}
