package main

import (
	"flag"
	"gotunel/lib"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

type ConfigTransport struct {
	TargetPort int
	TargetHost string
	ServerPort int
	ServerIp   string
	TcpOrUdp   string
	MinConns   int
	MaxConns   int
}

type ConfigAuth struct {
	Name   string
	Symkey string
}

type Config struct {
	Auth      *ConfigAuth
	Transport []*ConfigTransport
}

var logger *zap.SugaredLogger

func init() {
	logger = lib.GetLogger()
}

func main() {
	var path string
	var config Config
	flag.StringVar(&path, "c", "inner.toml", "config path")
	flag.Parse()
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		panic(err)
	}
	for _, transportConfig := range config.Transport {
		t := transport{minConns: transportConfig.MinConns,
			maxConns:   transportConfig.MaxConns,
			serverIp:   transportConfig.ServerIp,
			serverPort: transportConfig.ServerPort,
			targetPort: transportConfig.TargetPort,
			targetHost: transportConfig.TargetHost,
			tcpOrUdp:   transportConfig.TcpOrUdp,
			name:       lib.String2Byte16(config.Auth.Name),
			symkey:     lib.String2Byte16(config.Auth.Symkey),
		}
		go t.createCmdAndConn()
		go t.monitor()
	}
	select {}
}
