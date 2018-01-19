package manager

import (
	"net/http"

	"github.com/dearcode/crab/http/server"
	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/doodle/util"
)

type resource struct {
	ResourceID int64 `json:"resourceID"`
	RoleID     int64 `json:"roleID"`
}

// GET 根据条件查询管理组.
func (res *resource) GET(w http.ResponseWriter, r *http.Request) {
	u, err := session.User(r)
	if err != nil {
		log.Errorf("session.User error:%v, req:%v", errors.ErrorStack(err), r)
		response(w, Response{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	if err = util.DecodeRequestValue(r, res); err != nil {
		response(w, Response{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	if !u.IsAdmin && res.ResourceID == 0 {
		log.Errorf("%v resource id is 0, vars:%v", r.RemoteAddr, res)
		response(w, Response{Status: http.StatusBadRequest, Message: "resourceID is 0"})
		return
	}

	rs, err := rbacClient.GetResourceRoles(res.ResourceID)
	if err != nil {
		response(w, Response{Status: http.StatusInternalServerError, Message: err.Error()})
		log.Errorf("query vars:%v error:%s", res, errors.ErrorStack(err))
		return
	}
	log.Debugf("query:%+v, resource:%v", res, rs)
	server.SendData(w, rs)
}

type resourceInfo struct {
	ResourceID int64  `json:"resource_id" validate:"Required"`
	Sort       string `json:"sort"`
	Order      string `json:"order"`
	Offset     int    `json:"offset"`
	Limit      int    `json:"limit"`
}

// GET 获取资源信息.
func (ri *resourceInfo) GET(w http.ResponseWriter, r *http.Request) {
	if err := util.DecodeRequestValue(r, ri); err != nil {
		response(w, Response{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	rs, err := rbacClient.GetResource(ri.ResourceID)
	if err != nil {
		response(w, Response{Status: http.StatusInternalServerError, Message: err.Error()})
		log.Errorf("query vars:%v error:%s", ri, errors.ErrorStack(err))
		return
	}

	log.Debugf("query:%v, resource:%v", ri, rs)
	server.SendResponseData(w, rs)
}

//POST 关联角色
func (res *resource) POST(w http.ResponseWriter, r *http.Request) {
	if err := util.DecodeRequestValue(r, res); err != nil {
		response(w, Response{Status: 500, Message: err.Error()})
		return
	}

	id, err := rbacClient.PostRoleResource(res.RoleID, res.ResourceID)
	if err != nil {
		log.Errorf("RelationResourceRoleAdd error, vars:%v, err:%v", res, err)
		response(w, Response{Status: 500, Message: err.Error()})
		return
	}

	log.Debugf("add relation vars:%v, id:%d", res, id)

	response(w, Response{Data: id})
}

type resourceRole struct {
	ID         int64 `json:"id"`
	ResourceID int64 `json:"resourceID"`
	RoleID     int64 `json:"roleID"`
}

//DELETE 解除关联
func (rr *resourceRole) DELETE(w http.ResponseWriter, r *http.Request) {
	if err := util.DecodeRequestValue(r, rr); err != nil {
		response(w, Response{Status: 500, Message: err.Error()})
		return
	}

	if err := rbacClient.DeleteResourceRole(rr.ResourceID, rr.RoleID); err != nil {
		log.Errorf("DeleteResourceRole error, vars:%v, err:%v", rr, err)
		response(w, Response{Status: 500, Message: err.Error()})
		return
	}

	log.Debugf("del relation vars:%+v", rr)

	response(w, Response{})
}
