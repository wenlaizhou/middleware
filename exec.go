package middleware

import "os/exec"
import "fmt"
import "context"
import "time"
import "bytes"
import "io"

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
	return fmt.Sprintf("%s\n%s", outStr, errStr), err
}