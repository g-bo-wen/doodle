package repeater

import (
	"sync"
	"time"

	"github.com/zssky/log"

	"github.com/dearcode/doodle/repeater/config"
)

type errorEntry struct {
	Session string
	App     int64
	Iface   int64
	Info    string
	Time    time.Time
}

type entry struct {
	App   int64
	Iface int64
	Count int
	Time  int64
}

type ifaceEntry struct {
	apps map[int64]*entry
}

type statsCache struct {
	ies map[int64]*ifaceEntry
	ees []*errorEntry
	sync.Mutex
}

func newStatsCache() *statsCache {
	return &statsCache{ies: make(map[int64]*ifaceEntry)}
}

//failed 添加异常记录
func (s *statsCache) failed(id string, app, iface int64, log string) {
	s.Lock()
	defer s.Unlock()

	s.ees = append(s.ees, &errorEntry{id, app, iface, log, time.Now().Add(time.Hour * 8)})
}

//success 添加记录, 合并同一app调用同一接口的统计
func (s *statsCache) success(app, iface, tm int64) {
	s.Lock()
	defer s.Unlock()

	ie, ok := s.ies[iface]
	if !ok {
		ie = &ifaceEntry{apps: make(map[int64]*entry)}
		s.ies[iface] = ie
	}

	e, ok := ie.apps[app]
	if !ok {
		e = &entry{app, iface, 0, 0}
		ie.apps[app] = e
	}

	e.Count++
	e.Time += tm
	log.Debugf("new log:%+v", *e)
}

//entrys 读取统计信息, 并清理
func (s *statsCache) entrys() []entry {
	s.Lock()
	defer s.Unlock()

	var es []entry

	for ii, ie := range s.ies {
		for ai, e := range ie.apps {
			es = append(es, *e)
			delete(ie.apps, ai)
		}
		delete(s.ies, ii)
	}

	return es
}

//errorEntrys 异常日志, 并清理
func (s *statsCache) errorEntrys() []*errorEntry {
	s.Lock()
	defer s.Unlock()

	ees := s.ees
	s.ees = []*errorEntry{}
	return ees
}

func (s *statsCache) run() {
	t := time.NewTicker(time.Duration(config.Repeater.Cache.Timeout) * time.Second)
	for {
		<-t.C
		for _, e := range s.entrys() {
			if err := dc.insertStats(e.Iface, e.App, e.Count, e.Time); err != nil {
				log.Errorf("insertStats %v error:%v", e, err.Error())
			}
		}

		for _, e := range s.errorEntrys() {
			if err := dc.insertErrorStats(e.Session, e.Iface, e.App, e.Info, e.Time); err != nil {
				log.Errorf("insertErrorStats %v error:%v", e, err.Error())
			}
		}
	}

}
