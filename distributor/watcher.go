package distributor

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/dearcode/crab/http/client"
	"github.com/dearcode/crab/http/server"
	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/doodle/distributor/config"
	"github.com/dearcode/doodle/meta"
	"github.com/dearcode/doodle/meta/document"
	"github.com/dearcode/doodle/util"
	"github.com/dearcode/doodle/util/etcd"
)

const (
	apigatePrefix = "/api"
)

type watcher struct {
	etcd *etcd.Client
	apps map[string]meta.MicroAPP
	mu   sync.RWMutex
}

func newWatcher() (*watcher, error) {
	c, err := etcd.New(strings.Split(config.Distributor.ETCD.Hosts, ","))
	if err != nil {
		return nil, errors.Annotatef(err, config.Distributor.ETCD.Hosts)
	}

	return &watcher{etcd: c, apps: make(map[string]meta.MicroAPP)}, nil
}

func (w *watcher) start() {
	ec := make(chan clientv3.Event)

	for {
		go w.etcd.WatchPrefix(apigatePrefix, ec)
		for e := range ec {
			// /api/dbs/dbfree/handler/Fore/192.168.180.102/21638
			ss := strings.Split(string(e.Kv.Key), "/")
			if len(ss) < 4 {
				log.Errorf("invalid key:%s, event:%v", e.Kv.Key, e.Type)
				continue
			}

			if e.Type != clientv3.EventTypePut {
				log.Infof("ignore event:%+v", e)
				continue
			}

			name := strings.Join(ss[2:len(ss)-2], "/")

			app := meta.MicroAPP{}
			json.Unmarshal(e.Kv.Value, &app)
			w.register(name, app)
		}
	}
}

func (w *watcher) load() error {
	bss, err := w.etcd.List(apigatePrefix)
	if err != nil {
		log.Errorf("list %s error:%v", apigatePrefix, err)
		return errors.Annotatef(err, apigatePrefix)
	}

	for k, v := range bss {
		// k = /api/git.jd.com/dbs/faas_test_001/192.168.137.222/41596
		ss := strings.Split(k, "/")
		if len(ss) < 4 {
			log.Errorf("invalid key:%s", k)
			continue
		}

		name := strings.Join(ss[2:len(ss)-2], "/")
		app := meta.MicroAPP{}
		json.Unmarshal([]byte(v), &app)
		w.register(name, app)
	}

	return nil
}

type managerClient struct {
}

func (mc *managerClient) interfaceRegister(projectID int64, name, method, path, backend string, m document.Method) error {
	url := fmt.Sprintf("%sinterface/register/", config.Distributor.Manager.URL)
	req := struct {
		Name      string
		ProjectID int64
		Path      string
		Method    server.Method
		Backend   string
		Comment   string
		Attr      document.Method
	}{
		Name:      name,
		ProjectID: projectID,
		Path:      path,
		Backend:   backend,
		Method:    server.NewMethod(method),
		Comment:   m.Comment,
		Attr:      m,
	}

	resp := struct {
		Status  int
		Data    int
		Message string
	}{}

	if err := client.New(config.Distributor.Server.Timeout).PostJSON(url, nil, req, &resp); err != nil {
		return errors.Annotatef(err, url)
	}

	if resp.Status != 0 {
		return errors.New(resp.Message)
	}

	log.Debugf("register %+v success, id:%v", req, resp.Data)

	return nil
}

const (
	httpConnectTimeout = 60
)

func (w *watcher) parseDocument(backend string, app meta.MicroAPP) error {
	url := fmt.Sprintf("http://%s:%d/document/", app.Host, app.Port)
	buf, err := client.New(httpConnectTimeout).Get(url, nil, nil)
	if err != nil {
		return errors.Trace(err)
	}

	var doc map[string]document.Module
	log.Debugf("source:%s", buf)

	if err = json.Unmarshal(buf, &doc); err != nil {
		log.Errorf("Unmarshal doc:%s error:%v", buf, err)
		return errors.Annotatef(err, "%s", buf)
	}

	log.Debugf("doc:%+v", doc)

	projectID, err := parseProjectID(app.ServiceKey)
	if err != nil {
		log.Errorf("parseProjectID:%s error:%v", app.ServiceKey, err)
		return errors.Annotatef(err, app.ServiceKey)
	}

	mc := managerClient{}
	for ok, ov := range doc {
		for mk, mv := range ov.Methods {
			mc.interfaceRegister(projectID, ok+"_"+mk, mk, ov.URL, backend, mv)
		}
	}

	return nil
}

//register 到管理处添加接口, 肯定是多个Distributor同时上报的，所以添加操作要指定版本信息.
func (w *watcher) register(backend string, app meta.MicroAPP) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.apps[backend]; ok {
		log.Debugf("app:%+v exist", app)
		return
	}

	w.apps[backend] = app
	w.parseDocument(backend, app)
	log.Debugf("new backend:%s, app:%+v", backend, app)
	return
}

func (w *watcher) stop() {
	w.etcd.Close()
}

func parseProjectID(key string) (int64, error) {
	aesKey := []byte(config.Distributor.Server.SecretKey)
	aesKey = append(aesKey, []byte(strings.Repeat("\x00", 8-len(aesKey)%8))...)

	buf, err := util.AesDecrypt(key, aesKey)
	if err != nil {
		return 0, errors.Trace(err)
	}

	var id int64
	_, err = fmt.Sscanf(string(buf), "%x.", &id)
	if err != nil {
		return 0, errors.Trace(err)
	}

	return id, nil
}
