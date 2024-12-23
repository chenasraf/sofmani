package utils

import (
	"io"
	"os"
	"os/exec"

	"github.com/chenasraf/sofmani/logger"
)

func RunCmdPassThrough(bin string, args ...string) error {
	logger.Debug("Running command: %s %v", bin, args)
	cmd := exec.Command(bin, args...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	err := cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func RunCmdPassThroughChained(commands [][]string) error {
	for _, c := range commands {
		err := RunCmdPassThrough(c[0], c[1:]...)
		if err != nil {
			return err
		}
	}
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
