package config

import (
	"flag"

	"github.com/dearcode/crab/config"
)

var (
	// Manager 配置信息.
	Manager    managerConfig
	configPath = flag.String("c", "./config/manager.ini", "config file")
)

type serverConfig struct {
	Domain  string
	WebPath string
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

type ssoConfig struct {
	Domain string
	Key    string
}

type etcdConfig struct {
	Hosts string
}

type rbacConfig struct {
	Host  string
	Token string
}

type managerConfig struct {
	Server serverConfig
	DB     dbConfig
	ETCD   etcdConfig
	RBAC   rbacConfig
	SSO    ssoConfig
	Cache  cacheConfig
}

//Load 加载配置文件.
func Load() error {
	return config.LoadConfig(*configPath, &Manager)
}
