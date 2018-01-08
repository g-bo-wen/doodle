package ssh

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
)

func writeLogStream(reader io.ReadCloser) {
	r := bufio.NewReader(reader)
	for {
		line, _, err := r.ReadLine()
		fmt.Printf("line:%s, err:%v\n", line, err)
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Printf("read error:%v\n", err)
			return
		}
	}
}

func writeLogs(pid int, stdOut, stdErr io.ReadCloser) {
	go writeLogStream(stdOut)
	go writeLogStream(stdErr)
}

func writeSSHLogs(stdOut, stdErr io.Reader) {

	writeLogs(0, ioutil.NopCloser(stdOut), ioutil.NopCloser(stdErr))
}

func TestExecPipe(t *testing.T) {
	sc := NewSSHClient("192.168.180.104", 22, "root", "1qaz@WSX")
	if err := sc.ExecPipe("hostname", writeSSHLogs); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestUpload(t *testing.T) {
	src := "/home/tian/work/bin/bee"
	dest := "dbfree"

	sc := NewSSHClient("192.168.180.104", 22, "root", "1qaz@WSX")
	if err := sc.Upload(src, dest); err != nil {
		t.Fatalf(err.Error())
	}

}
