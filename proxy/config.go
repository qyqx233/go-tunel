package main

import (
	"fmt"
	"io/ioutil"

	"github.com/rs/zerolog/log"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Transport []*ConfigTransport `yaml:"Transport"`
}

type ConfigTransport struct {
	Addr       string `yaml:"Addr"`
	TargetHost string `yaml:"TargetHost"`
	TargetPort int    `yaml:"TargetPort"`
	UsePool    bool   `yaml:"UsePool"`
}

var config Config

func parseConfig(configPath string) {
	var data []byte
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	config.Transport = append(config.Transport, &ConfigTransport{})
	config.Transport = append(config.Transport, &ConfigTransport{})
	data, err = yaml.Marshal(&config)
	log.Debug().Str("data", string(data)).Msg("data")
	fmt.Println(string(data))
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
	log.Debug().Interface("ss", &config).Msg("config")
	// _, err := toml.DecodeFile(configPath, &config)
	// if err != nil {
	// 	panic(err)
	// }
}
