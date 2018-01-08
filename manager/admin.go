package manager

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/dearcode/crab/cache"
	"github.com/dearcode/crab/meta"
	"github.com/dearcode/crab/orm"
	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/doodle/manager/config"
)

type userDB struct {
	admins *cache.Cache
	res    *cache.Cache
	sync.RWMutex
}

func newUserDB() *userDB {
	return &userDB{
		admins: cache.NewCache(int64(config.Manager.Cache.Timeout)),
		res:    cache.NewCache(int64(config.Manager.Cache.Timeout)),
	}
}

//isAdmin 判断是不是管理员
func (u *userDB) isAdmin(email string) bool {
	u.RLock()
	if ok := u.admins.Get(email); ok != nil {
		u.RUnlock()
		log.Debugf("email:%v, cache:%v", email, ok.(bool))
		return ok.(bool)
	}
	u.RUnlock()

	u.Lock()
	defer u.Unlock()

	if ok := u.admins.Get(email); ok != nil {
		log.Debugf("retry email:%v, cache:%v", email, ok.(bool))
		return ok.(bool)
	}

	db, err := mdb.GetConnection()
	if err != nil {
		log.Errorf("get db connection error:%v", errors.ErrorStack(err))
		return false
	}
	defer db.Close()

	admin := struct {
		User  string `db:"user"`
		Email string `db:"email"`
	}{}

	if err = orm.NewStmt(db, "admin").Where("email='%s'", email).Query(&admin); err != nil {
		if errors.Cause(err) == meta.ErrNotFound {
			log.Debugf("%s not admin", email)
			u.admins.Add(email, false)
			return false
		}
		log.Errorf("orm query error:%v", errors.ErrorStack(err))
		return false
	}

	u.admins.Add(email, true)

	log.Debugf("%v admin:%v", email, admin)

	return true
}

//loadResource 查找用户权限
func (u *userDB) loadResource(i *userinfo) error {
	u.RLock()
	if res := u.res.Get(i.Email); res != nil {
		u.RUnlock()
		log.Debugf("userinfo:%v, resource cache:%v", i, res.([]int64))
		i.setResource(res.([]int64))
		return nil
	}
	u.RUnlock()

	u.Lock()
	defer u.Unlock()

	if res := u.res.Get(i.Email); res != nil {
		log.Debugf("userinfo:%v, resource cache:%v", i, res.([]int64))
		i.setResource(res.([]int64))
		return nil
	}

	res, err := rbacClient.GetUserResourceIDs(i.Email)
	if err != nil {
		return errors.Trace(err)
	}
	i.setResource(res)
	log.Debugf("userinfo:%+v", *i)
	return nil
}

// setResource 设置用户允许使用的资源列表
func (u *userinfo) setResource(res []int64) {
	if u.Res = res; len(res) == 0 {
		return
	}

	buf := bytes.NewBufferString("")
	for _, id := range res {
		buf.WriteString(fmt.Sprintf("%d,", id))
	}
	buf.Truncate(buf.Len() - 1)
	u.ResKey = buf.String()
}

//validate 权限验证
func (u *userinfo) assert(resID int64) error {
	if u.IsAdmin {
		return nil
	}
	for _, id := range u.Res {
		if resID == id {
			return nil
		}
	}
	log.Errorf("account:%+v, resourceID:%d", *u, resID)
	return fmt.Errorf("you don't have permission to access")
}
