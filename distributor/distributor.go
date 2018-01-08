package distributor

import (
	"net/http"

	"github.com/dearcode/crab/http/server"
	"github.com/juju/errors"
	"github.com/zssky/log"
)

type deployLogs struct {
	ID         int64
	DeployID   int64
	INFO       string
	CreateTime string `db_default:"auto"`
}

type distributorLogs struct {
	ID            int64
	DistributorID int64
	State         int
	PID           int
	INFO          string
	CreateTime    string `db_default:"now()"`
}

type distributor struct {
	ID         int64
	ProjectID  int64
	Project    project `db_table:"one"`
	State      int
	Server     string
	CreateTime string `db_default:"now()"`
}

//POST 编译并更新指定项目.
func (d *distributor) POST(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		ProjectID int64
	}{}

	if err := server.ParseJSONVars(r, &vars); err != nil {
		server.SendResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	t, err := newTask(vars.ProjectID)
	if err != nil {
		log.Errorf("newWorkspace error:%v", errors.ErrorStack(err))
		server.SendResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Debugf("newWorkspace:%+v", t.ID)

	server.SendResponseData(w, t.d.ID)

	go d.run(t)

	return
}

func (d *distributor) run(t *task) {
	if err := t.compile(); err != nil {
		log.Errorf("newWorkspace error:%v", errors.ErrorStack(err))
		return
	}

	if err := t.install(); err != nil {
		log.Errorf("newWorkspace error:%v", errors.ErrorStack(err))
		return
	}

}
