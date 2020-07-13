package main

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Transport []*ConfigTransport
}

type ConfigTransport struct {
	Addr       string
	TargetHost string
	TargetPort int
}

var config Config

func parseConfig(configPath string) {
	_, err := toml.DecodeFile(configPath, &config)
	logger.Info("%d", len(config.Transport))
	if err != nil {
		panic(err)
	}
}
