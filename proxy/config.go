package main

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Transport []*ConfigTransport
}

type ConfigTransport struct {
	ListenPort int
	TargetHost string
	TargetPort int
}

var config Config

func parseConfig(configPath string) {
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		panic(err)
	}
}
