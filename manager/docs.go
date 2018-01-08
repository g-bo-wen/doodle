package manager

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zssky/log"

	"github.com/dearcode/doodle/util"
)

type docs struct {
}

// GET get docs.
func (d *docs) GET(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		ProjectName   string `json:"projectName"`
		InterfaceName string `json:"interfaceName"`
		Sort          string `json:"sort"`
		Order         string `json:"order"`
		Page          int    `json:"offset"`
		Size          int    `json:"limit"`
	}{}

	if err := util.DecodeRequestValue(r, &vars); err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	var where string

	if vars.ProjectName != "" {
		if where != "" {
			where += " and "
		}
		where += fmt.Sprintf("project.name like '%%%s%%'", vars.ProjectName)
	}

	if vars.InterfaceName != "" {
		if where != "" {
			where += " and "
		}
		where += fmt.Sprintf("interface.name like '%%%s%%'", vars.InterfaceName)
	}

	if where != "" {
		where += " and "
	}
	where += fmt.Sprintf("interface.project_id = project.id and interface.state = 1")

	switch vars.Sort {
	case "InterfaceName":
		vars.Sort = "interface.name"
	case "ProjectName":
		vars.Sort = "project.name"
	}

	items := []struct {
		ProjectID     int64  `db:"project.id"`
		Path          string `db:"CONCAT(project.path,'/',interface.path)"`
		ProjectName   string `db:"project.name"`
		InterfaceID   int64  `db:"interface.id"`
		InterfaceName string `db:"interface.name"`
		User          string `db:"interface.user"`
		Email         string `db:"interface.email"`
	}{}

	total, err := query("project,interface", where, vars.Sort, vars.Order, vars.Page, vars.Size, &items)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	if len(items) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"total":0,"rows":[]}`))
		log.Debugf("project not found")
		return
	}

	buf, err := json.Marshal(items)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"total":%d, "rows":%s}`, total, buf)))
}
