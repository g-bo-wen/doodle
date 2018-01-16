package manager

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dearcode/crab/http/server"
	"github.com/dearcode/crab/orm"
	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/doodle/manager/config"
	"github.com/dearcode/doodle/meta"
	"github.com/dearcode/doodle/util/etcd"
)

type service struct {
}

const (
	etcdAPIPrefix = "/api"
)

func (s *service) GET(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		ProjectID int64 `json:"projectID" valid:"Required"`
	}{}

	if err := server.ParseURLVars(r, &vars); err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	db, err := mdb.GetConnection()
	if err != nil {
		log.Errorf("GetConnection error:%v", errors.ErrorStack(err))
		response(w, Response{Status: http.StatusInternalServerError, Message: err.Error()})
		return
	}
	defer db.Close()

	var p meta.Project

	if err = orm.NewStmt(db, "project").Where("id=%d", vars.ProjectID).Query(&p); err != nil {
		log.Errorf("query project:%d error:%v", vars.ProjectID, errors.ErrorStack(err))
		response(w, Response{Status: http.StatusInternalServerError, Message: err.Error()})
		return
	}

	// source http://git.jd.com/dbs/faas_test_001
	key := etcdAPIPrefix + p.Source[6:]
	e, err := etcd.New(config.Manager.ETCD.Hosts)
	if err != nil {
		log.Errorf("connect etcd:%v error:%v", config.Manager.ETCD.Hosts, errors.ErrorStack(err))
		response(w, Response{Status: http.StatusInternalServerError, Message: err.Error()})
		return
	}

	km, err := e.List(key)
	if err != nil {
		log.Errorf("list etcd key:%v error:%v", key, errors.ErrorStack(err))
		response(w, Response{Status: http.StatusInternalServerError, Message: err.Error()})
		return
	}

	var rows []meta.MicroAPP

	for _, v := range km {
		var a meta.MicroAPP
		json.Unmarshal([]byte(v), &a)
		rows = append(rows, a)
	}

	log.Debugf("project:%v service:%+v", vars.ProjectID, rows)
	server.SendData(w, rows)
}
