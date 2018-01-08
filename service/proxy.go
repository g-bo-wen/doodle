package service

import (
	"net/http"
	"reflect"

	"github.com/dearcode/crab/http/server"
	"github.com/zssky/log"
)

const (
	projectPath = "github.com/dearcode"
)

//proxy 转http请求为函数调用.
func proxy(w http.ResponseWriter, r *http.Request, m reflect.Method) {
	reqType := m.Type.In(1)
	respType := m.Type.In(2).Elem()

	reqVal := reflect.New(reqType)
	respVal := reflect.New(respType)

	header := reqVal.Elem().FieldByName("APIHeader")
	if header.IsValid() {
		header.FieldByName("Session").SetString(r.Header.Get("Session"))
		header.FieldByName("Request").Set(reflect.ValueOf(*r))
	}

	if err := server.ParseJSONVars(r, reqVal.Interface()); err != nil {
		server.SendErrorDetail(w, http.StatusBadRequest, nil, err.Error())
		return
	}

	argv := []reflect.Value{reflect.New(m.Type.In(0)).Elem(), reqVal.Elem(), respVal}
	m.Func.Call(argv)
	server.SendData(w, respVal.Interface())
}

func (s *Service) handler(w http.ResponseWriter, r *http.Request) {
	m, ok := s.router.get(r.Method, r.URL.Path)
	log.Debugf("m:%v, ok:%v, method:%v, path:%v", m, ok, r.Method, r.URL.Path)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	proxy(w, r, m)
}
