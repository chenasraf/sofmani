package utils

import (
	"io"
	"os"
	"os/exec"
	"runtime"
	"slices"

	"github.com/chenasraf/sofmani/logger"
)

func RunCmdPassThrough(env []string, bin string, args ...string) error {
	logger.Debug("Running command: %s %v", bin, args)
	cmd := exec.Command(bin, args...)
	cmd.Env = slices.Concat(os.Environ(), cmd.Env, env)
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

func RunCmdPassThroughChained(env []string, commands [][]string) error {
	for _, c := range commands {
		err := RunCmdPassThrough(env, c[0], c[1:]...)
		if err != nil {
			return err
		}
	}
	return nil
}

func RunCmdGetSuccess(env []string, bin string, args ...string) (error, bool) {
	cmd := exec.Command(bin, args...)
	cmd.Env = slices.Concat(os.Environ(), cmd.Env, env)
	err := cmd.Run()
	if err != nil {
		return nil, false
	}
	return nil, true
}

func RunCmdGetOutput(env []string, bin string, args ...string) ([]byte, error) {
	cmd := exec.Command(bin, args...)
	cmd.Env = slices.Concat(os.Environ(), cmd.Env, env)
	out, err := cmd.Output()
	return out, err
}

func GetShellWhich() string {
	switch runtime.GOOS {
	case "windows":
		return "where"
	case "linux", "darwin":
		return "which"
	}
	return ""
}

func GetOSShell() string {
	switch runtime.GOOS {
	case "windows":
		return "cmd"
	case "linux", "darwin":
		return "sh"
	}
	return ""
}

func GetOSShellArgs(cmd string) []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"/C", cmd}
	case "linux", "darwin":
		return []string{"-c", cmd}
	}
	return []string{}
}
