package outer

import (
	"flag"

	"github.com/qyqx233/go-tunel/outer/pub"
	"github.com/qyqx233/go-tunel/outer/rest"
	"github.com/rs/zerolog/log"
)

func Start() {
	var path string
	flag.StringVar(&path, "c", "outer.toml", "config path")
	log.Info().Msgf("读取配置%s", path)
	config = pub.ParseConfig(path)
	go rest.StartRest(config.HttpServer.BindAddr())
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
			AddIp:      ch.AddIp,
			Export:     ch.Export,
		}
		log.Info().Msgf("添加远程服务%s:%s:%d", h.IP, h.TargetHost, h.TargetPort)
		transportMng.add(&h)
	}
	log.Info().Msgf("初始化了%d个transport", len(transportMng.tl))
	proxySvrMng = newproxySvrMng(config.ProxyServer.MinPort,
		config.ProxyServer.MaxPort)
	_, err := newCmdServer(config.CmdServer.BindAddr())
	if err != nil {
		panic(err)
	}
	select {}
}
