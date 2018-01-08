package ssh

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/juju/errors"
	"github.com/pkg/sftp"
	"github.com/zssky/log"
	"golang.org/x/crypto/ssh"
)

//Client ssh客户端，支持scp的.
type Client struct {
	server string
	user   string
	passwd string
	config ssh.ClientConfig
}

//NewClient 创建ssh客户端.
func NewClient(host string, port int, user, passwd string) *Client {
	return &Client{
		server: fmt.Sprintf("%s:%d", host, port),
		user:   user,
		passwd: passwd,
		config: ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				ssh.Password(passwd),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	}
}

//Exec 执行命令并等待返回结果.
func (s *Client) Exec(cmd string) (string, error) {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 1<<16)
			// 获取所有goroutine的stacktrace, 如果只获取当前goroutine的stacktrace, 第二个参数需要为 `false`
			runtime.Stack(buf, true)
			log.Errorf("panic err:%v", err)
			log.Errorf("panic stack:%v", string(buf))
		}
	}()

	client, err := ssh.Dial("tcp", s.server, &s.config)
	if err != nil {
		return "", errors.Annotatef(err, s.server)
	}

	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", errors.Annotatef(err, s.server)
	}
	defer session.Close()

	var bufErr, bufOut bytes.Buffer

	session.Stdout = &bufOut
	session.Stderr = &bufErr

	if err = session.Run(cmd); err != nil {
		return "", errors.Annotatef(err, bufOut.String()+bufErr.String())
	}

	return strings.TrimSpace(bufOut.String() + bufErr.String()), nil
}

//ExecPipe 执行命令并设置输出流.
func (s *Client) ExecPipe(cmdStr string, setPipe func(stdOut, stdErr io.Reader)) error {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 1<<16)
			// 获取所有goroutine的stacktrace, 如果只获取当前goroutine的stacktrace, 第二个参数需要为 `false`
			runtime.Stack(buf, true)
			log.Errorf("panic err:%v", err)
			log.Errorf("panic stack:%v", string(buf))
		}
	}()

	client, err := ssh.Dial("tcp", s.server, &s.config)
	if err != nil {
		return errors.Annotatef(err, s.server)
	}

	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return errors.Annotatef(err, s.server)
	}
	defer session.Close()

	stdErr, err := session.StderrPipe()
	if err != nil {
		return errors.Trace(err)
	}

	stdOut, err := session.StdoutPipe()
	if err != nil {
		return errors.Trace(err)
	}

	setPipe(stdOut, stdErr)

	return errors.Trace(session.Run(cmdStr))
}

//Upload 上传文件.
func (s *Client) Upload(src, dest string) error {
	conn, err := ssh.Dial("tcp", s.server, &s.config)
	if err != nil {
		return errors.Annotatef(err, s.server)
	}
	sftp, err := sftp.NewClient(conn)
	if err != nil {
		return errors.Trace(err)
	}
	defer sftp.Close()

	st, err := os.Stat(src)
	if err != nil {
		return errors.Annotatef(err, "stat file:%v", src)
	}

	out, err := sftp.Create(dest)
	if err != nil {
		return errors.Annotatef(err, "create dest file:%v", dest)
	}
	defer out.Close()

	in, err := os.Open(src)
	if err != nil {
		return errors.Annotatef(err, "open src file:%v", src)
	}
	defer in.Close()

	if _, err = io.Copy(out, in); err != nil {
		return errors.Annotatef(err, "io copy src:%v, dest:%v", src, dest)
	}

	cmd := fmt.Sprintf("chmod %o %v", st.Mode(), dest)
	if _, err = s.Exec(cmd); err != nil {
		return errors.Annotatef(err, "cmd:%v", cmd)
	}

	return nil
}
