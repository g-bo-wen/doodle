package manager

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//userinfo erp中用户信息
type userinfo struct {
	ID        int64
	Status    int
	IsAdmin   bool
	Res       []int64
	ResKey    string
	Email     string `json:"email"`
	Mobile    string `json:"mobile"`
	HrmDeptID string `json:"hrmDeptId"`
	PersonID  string `json:"personId"`
	OrgID     string `json:"orgId"`
	OrgName   string `json:"orgName"`
	User      string `json:"fullname"`
	UserID    int    `json:"userId"`
	Username  string `json:"username"`
}

type statsSum struct {
	Date string
	Sum  int64
	Avg  int64
}

type statsTopApp struct {
	AppID   int64
	AppName string
	AppUser string

	InterfaceID   int64
	InterfaceName string
	InterfaceUser string

	ProjectID   int64
	ProjectName string

	Value int64
}

type statsTopIface struct {
	ID            int64  `json:"id"`
	ProjectName   string `json:"project"`
	InterfaceName string `json:"iface"`
	User          string `json:"user"`
	Value         int64  `json:"value"`
}

//ssoResp sso登录后返回的结构
type ssoResp struct {
	Code int      `json:"REQ_CODE"`
	Data userinfo `json:"REQ_DATA"`
	Flag bool     `json:"REQ_FLAG"`
	Msg  string   `json:"REQ_MSG"`
}

//Response 通用返回
type Response struct {
	Status  int
	Message string      `json:",omitempty"`
	Data    interface{} `json:",omitempty"`
}

//QueryResponse 专门给bootstrap-table用的.
type QueryResponse struct {
	Total int         `json:"total"`
	Rows  interface{} `json:"rows"`
}

func response(w http.ResponseWriter, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	buf, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(w, `{"Status":500, "Message":"%s"}`, err.Error())
		return
	}
	if _, err = w.Write(buf); err != nil {
		fmt.Fprintf(w, `{"Status":500, "Message":"%s"}`, err.Error())
	}
}

type iface struct {
	ID        int64  `db_default:"auto"`
	ProjectID int64  `json:"pid" valid:"Required"`
	Name      string `json:"name"  valid:"Required"`
	Method    int    `json:"method"`
	User      string `json:"user"`
	Email     string `json:"email"`
	State     int
	Path      string `json:"path"  valid:"AlphaNumeric"`
	Backend   string `json:"backend"  valid:"Required"`
	Comments  string `json:"comments"  valid:"Required"`
	Level     int    `json:"level"`
	CTime     string `db_default:"now()"`
	Mtime     string `db_default:"now()"`
}

type statsError struct {
	ID          int64
	Session     string
	AppID       int64  `db:"app_id"`
	AppName     string `db:"application.name"`
	IfaceID     int64  `db:"iface_id"`
	IfaceName   string `db:"interface.name"`
	ProjectName string `db:"project.name"`
	Info        string
	Ctime       string
}
