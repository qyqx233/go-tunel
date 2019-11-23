package main

import (
	"strconv"

	"github.com/BurntSushi/toml"
)

type Config struct {
	ProxyServer *ConfigProxy
	Transport   []*ConfigTransport
	CmdServer   *ConfigCmd
	HttpServer  *ConfigHttp
	Global      *ConfigGlobal
}

type ConfigGlobal struct {
	WaitNewConnSeconds int64
}

type ConfigTransport struct {
	IP         string
	TargetHost string
	TargetPort int
	Symkey     string
	MinConnNum int
	MaxConnNum int
	LocalPort  int
	KeepAlive  bool
}

type ConfigProxy struct {
	MinPort int
	MaxPort int
}

type ConfigServer struct {
	Addr string
	Port int
}

func (s ConfigServer) bindAddr() string {
	if s.Addr != "" {
		return s.Addr
	}
	return ":" + strconv.Itoa(s.Port)
}

type ConfigCmd struct {
	ConfigServer
}

type ConfigHttp struct {
	ConfigServer
}

var config Config

func parseConfig(configPath string) {
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		panic(err)
	}
	for _, host := range config.Transport {
		if len(host.Symkey) != 16 {
			panic("Symkey必须为16个字符")
		}
	}
}
