package distributor

import (
	"github.com/dearcode/crab/http/server"
	"github.com/dearcode/crab/orm"
	"github.com/juju/errors"

	"github.com/dearcode/doodle/distributor/config"
)

var (
	mdb *orm.DB
)

// ServerInit 初始化HTTP接口.
func ServerInit(confPath string) error {
	if err := config.Load(confPath); err != nil {
		return errors.Trace(err)
	}

	mdb = orm.NewDB(config.Distributor.DB.IP, config.Distributor.DB.Port, config.Distributor.DB.Name, config.Distributor.DB.User, config.Distributor.DB.Passwd, config.Distributor.DB.Charset, 10)

	server.RegisterPath(&distributor{}, "/distributor/")
	server.RegisterPath(&deploy{}, "/deploy/")

	w, err := newWatcher()
	if err != nil {
		return errors.Trace(err)
	}

	go w.start()

	if err = w.load(); err != nil {
		return errors.Trace(err)
	}

	return nil
}
