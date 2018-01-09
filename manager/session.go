package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dearcode/crab/cache"
	"github.com/juju/errors"

	"github.com/dearcode/doodle/manager/config"
)

const (
	//session会话超时30分钟
	sessionTimeout = 1800
)

type sessionDB struct {
	cache *cache.Cache
}

func newSessionDB() *sessionDB {
	return &sessionDB{cache: cache.NewCache(sessionTimeout)}
}

//getTicket 读取用户cookie中ticket
func getTicket(r *http.Request) (string, error) {
	c, err := r.Cookie(config.Manager.SSO.Key)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

//verifyTicket 调用sso接口验证ticket
func verifyTicket(r *http.Request, ticket string) (ssoResp, error) {
	ip := strings.Split(r.RemoteAddr, ":")[0]
	var sr ssoResp
	url := fmt.Sprintf("http://%s/sso/ticket/verifyTicket?ticket=%s&url=%s&ip=%s", config.Manager.SSO.Domain, ticket, r.URL.String(), ip)
	resp, err := http.Get(url)
	if err != nil {
		return sr, errors.Trace(err)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return sr, errors.Trace(err)
	}

	if err = json.Unmarshal(buf, &sr); err != nil {
		return sr, errors.Trace(err)
	}
	return sr, nil
}

func (u *userinfo) String() string {
	return u.Email
}

//loadInfo 加载资源与角色信息.
func (u *userinfo) loadInfo() error {
	res, err := userdb.loadResource(u.Email)
	if err != nil {
		return errors.Trace(err)
	}
	u.setResource(res)

	roles, err := userdb.loadRoles(u.Email)
	if err != nil {
		return errors.Trace(err)
	}

	u.setRoles(roles)

	u.IsAdmin = userdb.isAdmin(u.Email)

	return nil
}

func (s *sessionDB) User(r *http.Request) (*userinfo, error) {
	ticket, err := getTicket(r)
	if err != nil {
		return nil, errors.Trace(err)
	}

	val := s.cache.Get(ticket)
	if val != nil {
		i := val.(*userinfo)
		return i, i.loadInfo()
	}

	resp, err := verifyTicket(r, ticket)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if !resp.Flag {
		return nil, fmt.Errorf("invalid Flag, ticket:%+v", resp)
	}

	i := &resp.Data
	if err = i.loadInfo(); err != nil {
		return nil, errors.Trace(err)
	}

	s.cache.Add(ticket, i)

	return i, nil
}
