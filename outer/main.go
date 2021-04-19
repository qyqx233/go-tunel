package outer

import (
	"flag"
	"github.com/qyqx233/go-tunel/lib"
)

var logger = lib.GetLogger()

func Start() {
	var path string
	flag.StringVar(&path, "c", "outer.toml", "config path")
	parseConfig(path)
	for _, ch := range config.Transport {
		h := transportStru{
			IP:         ch.IP,
			TargetHost: ch.TargetHost,
			TargetPort: ch.TargetPort,
			SymKey:     []byte(ch.Symkey),
			MinConnNum: ch.MinConnNum,
			MaxConnNum: ch.MaxConnNum,
			LocalPort:  ch.LocalPort,
		}
		logger.Infof("添加远程服务%s:%s:%d", h.IP, h.TargetHost, h.TargetPort)
		transportMng.add(&h)
	}
	logger.Infof("初始化了%d个transport", len(transportMng.tl))
	proxySvrMng = newproxySvrMng(config.ProxyServer.MinPort,
		config.ProxyServer.MaxPort)
	go httpSvr(config.HttpServer.bindAddr())
	_, err := newCmdServer(config.CmdServer.bindAddr())
	if err != nil {
		panic(err)
	}
	select {}
}
