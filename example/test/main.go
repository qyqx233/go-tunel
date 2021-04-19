package main

import (
	"fmt"
	"github.com/qyqx233/go-tunel/pool"

	sentinel "github.com/alibaba/sentinel-golang/api"
)

var pools pool.Pool

func fastrand() uint32
func main() {
	var configPath = ""
	err := sentinel.Init(configPath)
	if err != nil {
		// 初始化 Sentinel 失败
	}
	fmt.Println(pools)
	// fmt.Println(fastrand())
	// reflect.MakeMapWithSize(int, 10)
}
