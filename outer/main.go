package outer

import (
	"flag"
	"sort"

	"github.com/qyqx233/go-tunel/outer/pub"
	"github.com/qyqx233/go-tunel/outer/rest"
	"github.com/rs/zerolog/log"
)

func initTransportMng() {
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
		// log.Info().Msgf("添加远程服务%s:%s:%d", h.IP, h.TargetHost, h.TargetPort)
		log.Info().Str("ip", h.IP).Str("targetHost", h.TargetHost).Str("key", ch.Symkey).Int("targetPort", h.TargetPort).Msg("添加远程服务")
		transportMng.add(&h)
		var tpdb rest.TransportPdb
		err := rest.PebbleGet(tpdb.GetKey(ch.LocalPort), &tpdb)
		if err != nil {
			pub.MemStor.Transports[ch.LocalPort] = &pub.Transport{
				Enable: tpdb.Enable,
			}
		} else {
			pub.MemStor.Transports[ch.LocalPort] = &pub.Transport{
				Enable: true,
			}
		}
	}
	sort.Slice(transportMng.tl, func(i, j int) bool {
		return transportMng.tl[i].TargetHost < transportMng.tl[j].TargetHost && transportMng.tl[i].TargetPort < transportMng.tl[j].TargetPort
	})
	log.Info().Msgf("初始化了%d个transport", len(transportMng.tl))
	log.Debug().Interface("tl", transportMng.tl).Msg("打印transportMng")
}

func Start() {
	var path string
	flag.StringVar(&path, "c", "outer.toml", "config path")
	log.Info().Msgf("读取配置%s", path)
	config = pub.ParseConfig(path)
	initTransportMng()
	go rest.StartRest(config.HttpServer.BindAddr())
	// 初始化结构体tranportMng，pub.MemStor（内存中节点元信息）
	proxySvrMng = newproxySvrMng(config.ProxyServer.MinPort,
		config.ProxyServer.MaxPort)
	_, err := newCmdServer(config.CmdServer.BindAddr())
	if err != nil {
		panic(err)
	}
	proxySvrMng.handle() // 监听server的管理
	select {}
}
