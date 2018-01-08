package config

import (
	"github.com/dearcode/crab/config"
)

var (
	// Distributor 配置信息.
	Distributor distributorConfig
)

type serverConfig struct {
	Timeout   int
	Domain    string
	Script    string
	BuildPath string
	SecretKey string
}

type dbConfig struct {
	IP      string
	Port    int
	Name    string
	User    string
	Passwd  string
	Charset string
}

type managerConfig struct {
	URL string
}

type etcdConfig struct {
	Hosts string
}

type distributorConfig struct {
	Server  serverConfig
	DB      dbConfig
	ETCD    etcdConfig
	Manager managerConfig
}

//Load 加载配置文件.
func Load(confPath string) error {
	return config.LoadConfig(confPath, &Distributor)
}
