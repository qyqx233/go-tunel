package outer

import (
	"github.com/BurntSushi/toml"
	"github.com/qyqx233/go-tunel/outer/pub"
)

var config *pub.Config

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
