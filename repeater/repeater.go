package repeater

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/dearcode/crab/http/server"
	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/doodle/meta"
	"github.com/dearcode/doodle/util"
	"github.com/dearcode/doodle/util/uuid"
)

func (r *repeater) delValue(req *http.Request, v *meta.Variable) {
	log.Debugf("del %v %v", v.Postion, v.Name)
	switch v.Postion {
	case server.URI:
		m := req.URL.Query()
		m.Del(v.Name)
	case server.HEADER:
		req.Header.Del(v.Name)
	case server.FORM:
		req.Form.Del(v.Name)
	}
}

func (r *repeater) validateValue(v *meta.Variable, val string) (bool, error) {
	if !v.Required && val == "" {
		return true, nil
	}

	if v.Type == "int" {
		_, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			log.Infof("val:%s, ParseInt error:%s, in %s", val, err.Error(), v.Postion)
			return false, fmt.Errorf("key:%s val:%s is not number, in %s", v.Name, val, v.Postion)
		}
		return false, nil
	}

	if val == "" {
		return false, fmt.Errorf("key:%s not found in %s", v.Name, v.Postion)
	}

	return false, nil
}

// GetInterface 根据请求header获取对应接口
func (r *repeater) GetInterface(req *http.Request, id string) (app *meta.Application, iface *meta.Interface, err error) {
	token := req.Header.Get("Token")
	if token == "" {
		return nil, nil, fmt.Errorf("need Token in Header")
	}

	log.Infof("%s requset token is:%v", id, token)

	if app, err = dc.getApp(token); err != nil {
		log.Errorf("%s get app error,token is:%v", id, token)
		return nil, nil, errors.Trace(err)
	}
	log.Infof("%s app is:%v, user email is:%v", id, app.Name, app.Email)

	if iface, err = dc.getInterface(req.URL.Path); err != nil {
		log.Errorf("%s get interface error path:%v, user email is:%v", id, req.URL.Path, app.Email)
		return nil, nil, errors.Trace(err)
	}
	log.Infof("%s iface is:%v,user email is:%v", id, iface.Path, iface.Email)

	if iface.Method != server.RESTful && req.Method != iface.Method.String() {
		log.Errorf("%s url:%v, invalid method:%v, need:%v,user email is:%v", id, req.URL, req.Method, iface.Method, iface.Email)
		return nil, nil, fmt.Errorf("invalid method:%v, need:%v", req.Method, iface.Method)
	}

	//如果不需要验证权限，直接通过
	if !iface.Service.Validate {
		log.Debugf("%s project:%v validate is flase, app:%v", id, id, iface.Service, app)
		return
	}

	if err = dc.validateRelation(app.ID, iface.ID); err == nil {
		log.Debugf("%s project:%v iface:%v app:%v", id, iface, app)
		return
	}

	if errors.Cause(err) == errNotFound {
		return nil, nil, errors.Trace(errForbidden)
	}

	log.Errorf("%s validateRelation appId:%v,ifaceId:%v,app email:%v,iface email is:%v", id, app.ID, iface.ID, app.Email, iface.Email)
	return nil, nil, errors.Trace(err)
}

func (r *repeater) parseForm(req *http.Request, vars []*meta.Variable) error {
	//如果需要解析body，则要备份一份
	for _, v := range vars {
		if v.Postion == server.FORM {
			buf, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return errors.Trace(err)
			}

			req.Body = ioutil.NopCloser(bytes.NewReader(buf))
			//执行这个操作会把body读空
			if err := req.ParseForm(); err != nil {
				return errors.Trace(err)
			}
			//再还回去
			req.Body = ioutil.NopCloser(bytes.NewReader(buf))
			break
		}
	}

	return nil
}

//Validate 验证输入参数，如果通过验证返回后端地址
func (r *repeater) Validate(req *http.Request, iface *meta.Interface) error {
	vars, err := dc.getVariable(iface.ID)
	if err != nil {
		return errors.Trace(err)
	}

	if err = r.parseForm(req, vars); err != nil {
		return errors.Trace(err)
	}

	for _, v := range vars {
		var val string
		switch v.Postion {
		case server.URI:
			val = req.URL.Query().Get(v.Name)
		case server.FORM:
			val = req.FormValue(v.Name)
		case server.HEADER:
			val = req.Header.Get(v.Name)
		}
		del, err := r.validateValue(v, val)
		if err != nil {
			return errors.Trace(err)
		}
		if del {
			r.delValue(req, v)
		}
	}

	return nil
}

//buildRequest 生成后端请求request,清理无用的请求参数
func (r *repeater) buildRequest(id string, iface *meta.Interface, req *http.Request) error {
	backend := iface.Backend

	if iface.Service.Version == 1 {
		apps, err := bs.getMicroAPPs(iface.Backend)
		if err != nil {
			return errors.Trace(err)
		}
		idx := time.Now().UnixNano() % int64(len(apps))
		backend = fmt.Sprintf("http://%s:%d%s", apps[idx].Host, apps[idx].Port, iface.Path)
		log.Infof("faas backend url:%v", backend)
	} else {
		//取url中接口名后剩余部分, 可能是RESTful请求
		if idx := strings.Index(req.URL.Path, iface.Path); idx > 0 {
			if path := req.URL.Path[idx+len(iface.Path):]; len(path) > 1 {
				if strings.HasSuffix(backend, "/") {
					backend += path[1:]
				} else {
					backend += path
				}
			}
		}
	}

	//生成url参数
	if args := req.URL.Query().Encode(); args != "" {
		backend += "?" + args
	}

	var err error
	if req.URL, err = url.Parse(backend); err != nil {
		return errors.Trace(err)
	}

	req.Host = req.URL.Host
	req.Header.Set("User-Agent", "APIGate "+util.GitTime)
	req.Header.Del("Token")
	req.RequestURI = ""
	req.Header.Set("Session", id)

	return nil
}

//ServeHTTP 入口
func (r *repeater) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	id := uuid.String()
	w.Header().Add("Session", id)

	defer func() {
		if e := recover(); e != nil {
			log.Errorf("%s recover %v", id, e)
			log.Errorf("%s", debug.Stack())
			util.SendResponse(w, http.StatusBadRequest, "%v", e)
		}
	}()

	log.Infof("%s url:%v method:%v", id, req.URL, req.Method)
	//接口接收到请求的详细信息
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Errorf("%s getBody error:%s, req:%v", id, errors.ErrorStack(err), *req)
		return
	}

	log.Infof("%s data:%v", id, string(buf))
	//再还回去
	req.Body = ioutil.NopCloser(bytes.NewReader(buf))

	//查找对应接口信息
	app, iface, err := r.GetInterface(req, id)
	if err != nil {
		if errors.Cause(err) == errForbidden {
			log.Infof("%s forbidden", id)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		log.Errorf("%s error:%s", id, errors.ErrorStack(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	log.Infof("%s app:%s interface:%s path:%s", id, app.Name, iface.Name, iface.Path)

	//验证输入参数
	if err = r.Validate(req, iface); err != nil {
		log.Errorf("%s app email:%v, iface email:%v validate error:%s", id, app.Email, iface.Email, errors.ErrorStack(err))
		util.SendResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Infof("%s validate success", id)

	//生成后端请求
	if err = r.buildRequest(id, iface, req); err != nil {
		log.Errorf("%s app email:%v, iface email:%v buildRequest error:%s", id, app.Email, iface.Email, errors.ErrorStack(err))
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Infof("%s backUrl:%s method:%s begin", id, req.URL, iface.Method)

	b := time.Now()
	rb, code, err := util.DoRequest(req)
	cost := time.Since(b) / time.Millisecond

	if err != nil {
		stats.failed(id, app.ID, iface.ID, err.Error())
		log.Errorf("%s app email:%v,iface email:%v millisecond:%d end error:%s", id, app.Email, iface.Email, cost, err.Error())
		util.SendResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if code != http.StatusOK {
		stats.failed(id, app.ID, iface.ID, fmt.Sprintf("invalid http status:%v", code))
		log.Errorf("%s app email:%v,iface email:%v millisecond:%d end failed, code:%d", id, app.Email, iface.Email, cost, code)
	} else {
		stats.success(app.ID, iface.ID, int64(cost))

		log.Infof("%s %d end success, code:%d", id, cost, code)
	}

	w.WriteHeader(code)
	w.Write(rb)
}
