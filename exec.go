package middleware

import (
	"io/ioutil"
	"os/exec"
	"strings"
)
import "fmt"
import "context"
import "time"
import "bytes"
import "io"

// 执行命令带超时时间
//
// timeout: 超时时间(单位秒)
func ExecCmdWithTimeout(timeout int, cmd string, args ...string) (string, error) {
	var err error
	var cmdHandler *exec.Cmd
	var arguments []string
	for _, c := range args {
		arguments = append(arguments, c)
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	cmdHandler = exec.CommandContext(ctx, cmd, arguments...)
	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmdHandler.StdoutPipe()
	stderrIn, _ := cmdHandler.StderrPipe()
	err = cmdHandler.Start()
	if err != nil {
		return "", err
	}
	_, _ = io.Copy(&stdoutBuf, stdoutIn)
	_, _ = io.Copy(&stderrBuf, stderrIn)
	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	_ = cmdHandler.Wait()
	return strings.TrimSpace(fmt.Sprintf("%s\n%s",
			strings.TrimSpace(outStr),
			strings.TrimSpace(errStr))),
		err
}

var interExecChannel map[string]chan string

// 交互式命令行服务
func InterExec() chan string {
	cmd := exec.Command("sh")
	w, _ := cmd.StdinPipe()
	o, _ := cmd.StdoutPipe()
	channel := make(chan string)
	go func() {
		var b bytes.Buffer
		cmd.Stdout = &b
		cmd.Stderr = &b
		cmd.Start()
		t := time.Now()
		for {
			select {
			case cmd := <-channel:
				t = time.Now()
				_, _ = w.Write([]byte(cmd))
				res, _ := ioutil.ReadAll(o)
				channel <- string(res)
				break
			default:
				if time.Now().Sub(t) > time.Minute*30 {
					println("channel timeout")
					close(channel)
					cmd.Process.Kill()
					return
				}
				break
			}
		}
	}()
	return channel
}
