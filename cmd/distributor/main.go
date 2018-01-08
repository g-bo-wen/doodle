package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dearcode/crab/http/server"
	"github.com/zssky/log"

	"github.com/dearcode/doodle/distributor"
	"github.com/dearcode/doodle/util"
)

var (
	addr       = flag.String("h", ":9300", "api listen address")
	debug      = flag.Bool("debug", false, "debug write log to console.")
	version    = flag.Bool("v", false, "show version info")
	configPath = flag.String("c", "./config/distributor.ini", "config file")

	maxWaitTime = time.Minute
)

func main() {
	flag.Parse()

	if *version {
		util.PrintVersion()
		return
	}

	if !*debug {
		log.SetOutputByName("./logs/api.log")
		log.SetHighlighting(false)
		log.SetRotateByDay()
	}

	if err := distributor.ServerInit(*configPath); err != nil {
		panic(err)
	}

	ln, err := server.Start(*addr)
	if err != nil {
		panic(err)
	}

	log.Infof("listener %s", ln.Addr())

	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, syscall.SIGUSR1)

	s := <-shutdown
	log.Warningf("recv signal %v, close.", s)
	ln.Close()
	time.Sleep(maxWaitTime)
	log.Warningf("server exit")
}
