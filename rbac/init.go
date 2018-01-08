package rbac

import (
	"encoding/binary"
	"net/http"

	"github.com/dearcode/crab/http/server"
	"github.com/dearcode/crab/orm"
	"github.com/juju/errors"

	"github.com/dearcode/doodle/rbac/config"
	"github.com/dearcode/doodle/util"
)

var (
	errInvalidToken = errors.New("invalid token")
	mdb             *orm.DB
)

// ServerInit 初始化HTTP接口.
func ServerInit() error {
	if err := config.Load(); err != nil {
		return err
	}

	mdb = orm.NewDB(config.RBAC.DB.IP, config.RBAC.DB.Port, config.RBAC.DB.Name, config.RBAC.DB.User, config.RBAC.DB.Passwd, config.RBAC.DB.Charset, 10)

	server.RegisterPath(&rbacUser{}, "/rbac/user/")
	server.RegisterPath(&userRole{}, "/rbac/user/role/")
	server.RegisterPath(&rbacRoleUser{}, "/rbac/role/user/")
	server.RegisterPath(&rbacRoleResource{}, "/rbac/role/resource/")
	server.RegisterPath(&resourceRolesUnrelated{}, "/rbac/resource/role/unrelated/")
	server.RegisterPath(&rbacRole{}, "/rbac/role/")
	server.RegisterPath(&rbacAPP{}, "/rbac/app/")
	server.RegisterPath(&rbacResource{}, "/rbac/resource/")
	server.RegisterPath(&userResource{}, "/rbac/user/resource/")

	return nil
}

func parseToken(r *http.Request) (int64, error) {
	token := r.Header.Get("token")
	buf, err := util.AesDecrypt(token, util.AesKey)
	if err != nil {
		return 0, errors.Annotatef(errInvalidToken, "token:%v, error:%v", token, err.Error())
	}

	id, n := binary.Varint(buf)
	if n < 1 {
		return 0, errors.Errorf("invalid token %s", token)
	}

	return id, nil
}
