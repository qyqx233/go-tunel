package inner

import (
	"flag"
	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
	"gotunel/lib"
	"sync"
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

var logger *zap.SugaredLogger

func init() {
	logger = lib.GetLogger()
}

func Start() {
	var path string
	var config Config
	flag.StringVar(&path, "c", "inner.toml", "config path")
	flag.Parse()
	logger.Debug(path)
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
	wg.Wait()
}
