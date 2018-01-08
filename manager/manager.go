package manager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/doodle/manager/config"
	"github.com/dearcode/doodle/util"
)

var (
	userdb  = newUserDB()
	session = newSessionDB()
)

type account struct {
}

//GET 获取用户帐号信息
func (a *account) GET(w http.ResponseWriter, r *http.Request) {
	i, err := session.User(r)
	if err != nil {
		log.Errorf("get session user error:%v", errors.ErrorStack(err))
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	buf, err := json.Marshal(i)
	if err != nil {
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(buf)
}

type domain struct {
}

//onDomainGet 获取配置文件中域名
func (d *domain) GET(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(config.Manager.Server.Domain))
}

type static struct {
}

//GET 静态文件
func (s *static) GET(w http.ResponseWriter, r *http.Request) {
	//	w.Header().Add("Cache-control", "no-store")
	path := fmt.Sprintf("%s%s", config.Manager.Server.WebPath, r.URL.Path)
	http.ServeFile(w, r, path)
}

type debug struct {
}

func (d *debug) GET(w http.ResponseWriter, r *http.Request) {
	pprof.Index(w, r)
}
