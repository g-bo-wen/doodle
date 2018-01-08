package config

import (
	"flag"

	"github.com/dearcode/crab/config"
)

var (
	// Repeater 配置信息.
	Repeater   repeaterConfig
	configPath = flag.String("c", "./config/repeater.ini", "config file")
)

type serverConfig struct {
	Timeout   int
	SecretKey string
}

type cacheConfig struct {
	Timeout int
}

type dbConfig struct {
	IP      string
	Port    int
	Name    string
	User    string
	Passwd  string
	Charset string
}

type etcdConfig struct {
	Hosts string
}

type repeaterConfig struct {
	Cache  cacheConfig
	Server serverConfig
	DB     dbConfig
	ETCD   etcdConfig
}

//Load 加载配置文件.
func Load() error {
	return config.LoadConfig(*configPath, &Repeater)
}
