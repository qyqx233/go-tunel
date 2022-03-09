package inner

import (
	"flag"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/qyqx233/go-tunel/lib"
)

type FatalError struct {
	code int8
	msg  string
}

func (f FatalError) Error() string {
	return f.msg
}

type ConfigTransport struct {
	TargetPort int
	TargetHost string
	ServerPort int
	ServerIp   string
	TcpOrUdp   string
	MinConns   int
	MaxConns   int
	ShakeRetry int
}

type ConfigAuth struct {
	Name   string
	Symkey string
}

type Config struct {
	Auth      *ConfigAuth
	Transport []*ConfigTransport
}

func Start() {
	var path string
	var config Config
	flag.StringVar(&path, "c", "inner.toml", "config path")
	flag.Parse()
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		panic(err)
	}
	wg := &sync.WaitGroup{}
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
			shakeRetry: transportConfig.ShakeRetry,
		}
		wg.Add(1)
		go t.createCmdAndConn()
		go t.monitor(wg)
	}
	select {}
}
