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

func (s *sessionDB) User(r *http.Request) (*userinfo, error) {
	ticket, err := getTicket(r)
	if err != nil {
		return nil, errors.Trace(err)
	}

	val := s.cache.Get(ticket)
	if val != nil {
		i := val.(*userinfo)
		if err = userdb.loadResource(i); err != nil {
			return nil, errors.Trace(err)
		}
		return i, nil
	}

	resp, err := verifyTicket(r, ticket)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if !resp.Flag {
		return nil, fmt.Errorf("invalid Flag, ticket:%+v", resp)
	}

	if err = userdb.loadResource(&resp.Data); err != nil {
		return nil, errors.Trace(err)
	}

	if resp.Data.UserID, err = userdb.loadUserID(resp.Data.Email); err != nil {
		return nil, errors.Trace(err)
	}

	resp.Data.IsAdmin = userdb.isAdmin(resp.Data.Email)

	s.cache.Add(ticket, &resp.Data)

	return &resp.Data, nil
}
