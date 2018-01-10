package service

import (
	"flag"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/doodle/meta"
	"github.com/dearcode/doodle/service/debug"
	"github.com/dearcode/doodle/util"
	"github.com/dearcode/doodle/util/etcd"
)

const (
	apigatePrefix = "/api/"
)

var (
	etcdAddrs = flag.String("etcd", "", "etcd Endpoints, like 192.168.180.104:12379,192.168.180.104:22379,192.168.180.104:32379.")
)

type keepalive struct {
	etcd *etcd.Client
}

func newKeepalive() *keepalive {
	if *etcdAddrs == "" {
		return nil
	}

	//清理输入ip列表.
	addrs := strings.Split(*etcdAddrs, ",")
	for i := range addrs {
		addrs[i] = strings.TrimSpace(addrs[i])
	}

	//连接etcd.
	c, err := etcd.New(addrs)
	if err != nil {
		panic(err.Error())
	}

	return &keepalive{etcd: c}
}

// register 服务上线，注册到接口平台的etcd.
func (k *keepalive) start(ln net.Listener, doc document) error {
	if k == nil {
		return nil
	}

	// 获取本机服务地址
	local := util.LocalAddr()

	la := ln.Addr().String()
	port := la[strings.LastIndex(la, ":")+1:]

	//注册服务，key为当前项目名及IP端口
	key := apigatePrefix + debug.Project + "/" + local + "/" + port
	p, _ := strconv.Atoi(port)
	val := meta.NewMicroAPP(debug.GitHash, local, debug.ServiceKey, p, os.Getpid()).String()

	if _, err := k.etcd.Keepalive(key, val); err != nil {
		log.Errorf("etcd Keepalive key:%v, val:%v, error:%v", key, val, errors.ErrorStack(err))
		return errors.Trace(err)
	}

	log.Debugf("etcd put key:%v val:%v", key, val)

	return nil
}

func (k *keepalive) stop() {
	if k == nil {
		return
	}

	k.etcd.Close()
}
