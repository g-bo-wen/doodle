package meta

import (
	"encoding/json"
)

//MicroAPP 一个函数式应用.
type MicroAPP struct {
	ServiceKey string
	Host       string
	Port       int
	PID        int
	GitHash    string
	GitTime    string
	GitMessage string
}

//NewMicroAPP 一个应用.
func NewMicroAPP(host, key string, port, pid int, hash, time, message string) *MicroAPP {
	return &MicroAPP{
		ServiceKey: key,
		PID:        pid,
		Host:       host,
		Port:       port,
		GitHash:    hash,
		GitTime:    time,
		GitMessage: message,
	}
}

func (m *MicroAPP) String() string {
	b, _ := json.Marshal(m)
	return string(b)
}
