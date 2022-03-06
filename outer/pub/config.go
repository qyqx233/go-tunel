package pub

import (
	"strconv"

	"github.com/BurntSushi/toml"
)

type Transport struct {
	Enable bool
}

type MemStorStru struct {
	Ips        map[string]struct{} `json:"ips"`
	Transports map[int]Transport   `json:"transports"`
}

var MemStor MemStorStru

type Config struct {
	ProxyServer *ConfigProxy
	Transport   []*ConfigTransport
	CmdServer   *ConfigCmd
	HttpServer  *ConfigHttp
	Global      *ConfigGlobal
	portTsMap   map[int]*ConfigTransport
}

func (c *Config) GetTs(port int) *ConfigTransport {
	return c.portTsMap[port]
}

type ConfigGlobal struct {
	WaitNewConnSeconds int64
}

type ConfigTransport struct {
	Dump       bool
	AddIp      bool
	Enable     bool
	Export     bool
	KeepAlive  bool
	TargetPort int
	MinConnNum int
	MaxConnNum int
	LocalPort  int
	Symkey     string
	IP         string
	TargetHost string
}

type ConfigProxy struct {
	MinPort int
	MaxPort int
}

type ConfigServer struct {
	Addr string
	Port int
}

func (s ConfigServer) BindAddr() string {
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

var config *Config

func GetConfig() *Config {
	return config
}

func ParseConfig(configPath string) *Config {
	if config != nil {
		return config
	}
	config = new(Config)
	config.portTsMap = make(map[int]*ConfigTransport)
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		panic(err)
	}
	for _, ts := range config.Transport {
		if len(ts.Symkey) != 16 {
			panic("Symkey必须为16个字符")
		}
		config.portTsMap[ts.LocalPort] = ts
	}
	return config
}

func init() {
	MemStor.Ips = make(map[string]struct{})
	MemStor.Transports = make(map[int]Transport)
}
