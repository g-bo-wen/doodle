package config

import (
	"flag"

	"github.com/dearcode/crab/config"
)

var (
	// RBAC 配置信息.
	RBAC       rbacConfig
	configPath = flag.String("c", "./config/rbac.ini", "config file")
)

type serverConfig struct {
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

type rbacConfig struct {
	Server serverConfig
	DB     dbConfig
}

//Load 加载配置文件.
func Load() error {
	return config.LoadConfig(*configPath, &RBAC)
}
