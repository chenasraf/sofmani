package utils

import (
	"io"
	"os"
	"os/exec"
)

func RunCmdPassThrough(bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	cmd.Wait()
	return nil
}

func RunCmdGetSuccess(bin string, args ...string) (error, bool) {
	cmd := exec.Command(bin, args...)
	err := cmd.Run()
	if err != nil {
		return nil, false
	}
	return nil, true
}

func RunCmdGetOutput(bin string, args ...string) ([]byte, error) {
	cmd := exec.Command(bin, args...)
	out, err := cmd.Output()
	return out, err
}
