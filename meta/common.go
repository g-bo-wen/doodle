package meta

import (
	"github.com/dearcode/crab/http/server"
)

//Application 对应应用表.
type Application struct {
	ID      int64
	Name    string
	User    string
	Email   string
	Token   string
	Comment string
	Ctime   string
	Mtime   string
}

//Relation 关联关系结构.
type Relation struct {
	ID               int64
	ApplicationID    int64  `db:"application_id"`
	ApplicationName  string `db:"application.name"`
	ApplicationUser  string `db:"application.user"`
	ApplicationEmail string `db:"application.email"`
	ProjectID        int64  `db:"project.id"`
	ProjectName      string `db:"project.name"`
	ProjectUser      string `db:"project.user"`
	ProjectEmail     string `db:"project.email"`
	InterfaceID      int64  `db:"interface.id"`
	InterfaceName    string `db:"interface.name"`
	Ctime            string
	Mtime            string
}

//Project 项目信息
type Project struct {
	ID         int64  `json:"id" db_default:"auto"`
	RoleID     int64  `json:"role_id" `
	ResourceID int64  `json:"resource_id" `
	Name       string `json:"name" valid:"Required"`
	User       string `json:"user" `
	Validate   bool
	Email      string `json:"email" `
	Path       string `json:"path"  valid:"AlphaNumeric"`
	Source     string `json:"source" `
	Version    int    `json:"version" `
	Comment    string `json:"comment" valid:"Required"`
	CTime      string `json:"ctime" db:"ctime" db_default:"now()"`
	MTime      string `json:"mtime" db:"mtime" db_default:"now()"`
}

//Variable 接口参数
type Variable struct {
	ID       int64
	Postion  server.VariablePostion
	Name     string
	Type     string
	Level    int
	Required bool
	Example  string
	Comment  string
	Ctime    string
	Mtime    string
}

// Interface 接口信息
type Interface struct {
	ID      int64
	Name    string
	User    string
	Email   string
	State   bool
	Project Project `db_table:"one"`
	Path    string
	Method  server.Method
	Backend string
	Comment string
	Level   int8
	Ctime   string
	Mtime   string
}

//TokenBody token结构.
type TokenBody struct {
	AppID      int64
	Name       string
	CreateTime int64
}

//Response 通用返回结果
type Response struct {
	Status  int
	Message string      `json:",omitempty"`
	Data    interface{} `json:",omitempty"`
}
