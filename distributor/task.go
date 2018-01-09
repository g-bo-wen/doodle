package distributor

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/dearcode/crab/orm"
	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/doodle/distributor/config"
	"github.com/dearcode/doodle/util"
	"github.com/dearcode/doodle/util/ssh"
	"github.com/dearcode/doodle/util/uuid"
)

const (
	stateInit = iota
	stateCompileBegin
	stateComplieSuccess
	stateComplieFailed
	stateInstallBegin
	stateInstallSuccess
	stateInstallFailed

	sqlWriteLogs       = "update distributor_logs set info = concat(info, ?) , state = ? where id=?"
	sqlWriteServerInfo = "update distributor set server = ?, pid = ? where id=?"
	sqlUpdateState     = "update distributor set state = ? where id=?"
)

var (
	scripts = []string{"build.sh", "Dockerfile.tpl"}
)

type task struct {
	wg      sync.WaitGroup
	ID      string
	path    string
	db      *sql.DB
	project project
	d       distributor
	state   int
	logID   int64
}

func newTask(projectID int64) (*task, error) {
	var p project

	db, err := mdb.GetConnection()
	if err != nil {
		return nil, errors.Trace(err)
	}

	if err = orm.NewStmt(db, "project").Where("project.id=%v", projectID).Query(&p); err != nil {
		return nil, errors.Trace(err)
	}
	log.Debugf("project:%#v", p)

	path := fmt.Sprintf("%s/%v", config.Distributor.Server.BuildPath, time.Now().UnixNano())
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, errors.Annotatef(err, path)
	}

	for _, f := range scripts {
		of := fmt.Sprintf("%s/%v", config.Distributor.Server.Script, f)
		nf := fmt.Sprintf("%s/%s", path, f)
		if err = os.Link(of, nf); err != nil {
			return nil, errors.Annotatef(err, "old:%v, new:%v", of, nf)
		}
	}

	d := distributor{
		ProjectID: projectID,
		Server:    util.LocalAddr(),
	}

	if d.ID, err = orm.NewStmt(db, "distributor").Insert(&d); err != nil {
		log.Errorf("insert distributor:%v error:%v", d, err)
		return nil, errors.Trace(err)
	}

	return &task{db: db, project: p, d: d, ID: uuid.String(), path: path}, nil
}

func (t *task) updateState(state int) {
	t.state = state
	if _, err := orm.NewStmt(t.db, "").Exec(sqlUpdateState, state, t.d.ID); err != nil {
		log.Errorf("%v sqlUpdateState:%v, %v, %v, error:%v", t, sqlUpdateState, state, t.d.ID, errors.ErrorStack(err))
		return
	}
}

func (t *task) String() string {
	return t.ID
}

func (t *task) writeLogStream(reader io.ReadCloser) {
	defer t.wg.Done()

	r := bufio.NewReader(reader)
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Errorf("%v read error:%v", t, err)
			return
		}
		log.Infof("%v %s", t, line)
		if _, err := orm.NewStmt(t.db, "").Exec(sqlWriteLogs, string(line)+"\n", t.state, t.logID); err != nil {
			log.Errorf("%v update logs sql:%v, %s, %v, error:%v", t, sqlWriteLogs, line, t.logID, err)
			continue
		}
	}
}

func (t *task) writeSSHLogs(stdOut, stdErr io.Reader) {
	t.writeLogs(0, ioutil.NopCloser(stdOut), ioutil.NopCloser(stdErr))
}

func (t *task) writeLogs(pid int, stdOut, stdErr io.ReadCloser) {
	dl := distributorLogs{
		DistributorID: t.d.ID,
		PID:           pid,
		State:         t.state,
	}

	id, err := orm.NewStmt(t.db, "distributor_logs").Insert(&dl)
	if err != nil {
		log.Errorf("%v insert distributor_logs error:%v", t, errors.ErrorStack(err))
		return
	}

	t.logID = id

	t.wg.Add(2)
	go t.writeLogStream(stdOut)
	go t.writeLogStream(stdErr)
}

func execSystemCmdWait(cmdStr string, stdPipe func(pid int, out, err io.ReadCloser)) error {
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return errors.Trace(err)
	}
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Trace(err)
	}

	if err := cmd.Start(); err != nil {
		return errors.Trace(err)
	}

	stdPipe(cmd.Process.Pid, outPipe, errPipe)

	return errors.Trace(cmd.Wait())
}

func (t *task) install() error {
	var err error

	t.updateState(stateInstallBegin)

	defer func() {
		if err != nil {
			t.updateState(stateInstallFailed)
			return
		}
		t.updateState(stateInstallSuccess)
	}()

	//切换工作目录.
	oldPath, _ := os.Getwd()
	if err = os.Chdir(t.path); err != nil {
		return errors.Annotatef(err, t.path)
	}
	defer os.Chdir(oldPath)

	tarFile := t.project.Name + ".tar.gz"

	cmd := fmt.Sprintf("tar -C bin -czf %s %s", tarFile, t.project.Name)
	if err := execSystemCmdWait(cmd, t.writeLogs); err != nil {
		return errors.Annotatef(err, cmd)
	}

	cmd = fmt.Sprintf("hostname; tar xzf %s; killall %v; nohup ./%v -etcd 192.168.180.104:12379,192.168.180.104:22379,192.168.180.104:32379 -h : > %v.log 2>&1 &", tarFile, t.project.Name, t.project.Name, t.project.Name)

	for _, n := range t.project.Cluster.Node {
		sc, err := ssh.NewClient(n.Server, 22, "jeduser", "", config.Distributor.SSH.Key)
		if err != nil {
			return errors.Trace(err)
		}
		log.Debugf("%v begin, upload file:%v", t, tarFile)
		if err = sc.Upload(tarFile, tarFile); err != nil {
			return errors.Trace(err)
		}
		log.Debugf("%v end, upload file:%v", t, tarFile)

		log.Debugf("%v begin, ssh exec:%v", t, cmd)
		if err = sc.ExecPipe(cmd, t.writeSSHLogs); err != nil {
			return errors.Annotatef(err, cmd)
		}
		log.Debugf("%v end, ssh exec:%v", t, cmd)
		log.Debugf("%v deploy %s success", t, n.Server)
	}
	t.wg.Wait()

	log.Debugf("%v deploy all success", t)

	return nil
}

//compile 使用脚本编译指定应用.
func (t *task) compile() error {
	var err error

	t.updateState(stateCompileBegin)

	defer func() {
		if err != nil {
			t.updateState(stateComplieFailed)
			return
		}
		t.updateState(stateComplieSuccess)
	}()

	oldPath, _ := os.Getwd()
	if err = os.Chdir(t.path); err != nil {
		return errors.Annotatef(err, t.path)
	}
	defer os.Chdir(oldPath)

	cmd := fmt.Sprintf("./build.sh %s %s", t.project.Source, t.project.key())

	if err = execSystemCmdWait(cmd, t.writeLogs); err != nil {
		return errors.Annotatef(err, cmd)
	}
	t.wg.Wait()

	log.Debugf("path:%v", oldPath)

	return nil
}
