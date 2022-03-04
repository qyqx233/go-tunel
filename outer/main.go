package outer

import (
	"flag"

	"github.com/rs/zerolog/log"
)

func Start() {
	var path string
	flag.StringVar(&path, "c", "outer.toml", "config path")
	parseConfig(path)
	for _, ch := range config.Transport {
		h := transportImpl{
			IP:         ch.IP,
			TargetHost: ch.TargetHost,
			TargetPort: ch.TargetPort,
			SymKey:     []byte(ch.Symkey),
			MinConnNum: ch.MinConnNum,
			MaxConnNum: ch.MaxConnNum,
			LocalPort:  ch.LocalPort,
			Dump:       ch.Dump,
		}
		log.Info().Msgf("添加远程服务%s:%s:%d", h.IP, h.TargetHost, h.TargetPort)
		transportMng.add(&h)
	}
	log.Info().Msgf("初始化了%d个transport", len(transportMng.tl))
	proxySvrMng = newproxySvrMng(config.ProxyServer.MinPort,
		config.ProxyServer.MaxPort)
	go httpSvr(config.HttpServer.bindAddr())
	_, err := newCmdServer(config.CmdServer.bindAddr())
	if err != nil {
		panic(err)
	}
	select {}
}
