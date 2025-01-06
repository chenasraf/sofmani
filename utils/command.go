package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/chenasraf/sofmani/logger"
)

const UNIX_DEFAULT_SHELL = "bash"

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

func getShellScript(dir string) string {
	var filename string
	switch runtime.GOOS {
	case "windows":
		filename = "install.bat"
	case "linux", "darwin":
		filename = "install"
	}
	tmpfile := filepath.Join(dir, filename)
	return tmpfile
}

func getScriptContents(script string, envShell *string) (string, error) {
	switch runtime.GOOS {
	case "windows":
		return script, nil
	case "linux", "darwin":
		if envShell == nil {
			shell := UNIX_DEFAULT_SHELL
			envShell = &shell
		}
		return fmt.Sprintf("#!/usr/bin/env %s\n%s\n", *envShell, script), nil
	}
	return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func RunCmdAsFile(env []string, contents string, envShell *string) error {
	tmpdir := os.TempDir()
	tmpfile := getShellScript(tmpdir)
	commandStr, err := getScriptContents(contents, envShell)
	if err != nil {
		return err
	}
	err = os.WriteFile(tmpfile, []byte(commandStr), 0755)
	if err != nil {
		return err
	}

	shell := GetOSShell(envShell)
	args := GetOSShellArgs(tmpfile)
	return RunCmdPassThrough(env, shell, args...)
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

func GetOSShell(envShell *string) string {
	switch runtime.GOOS {
	case "windows":
		return "cmd"
	case "linux", "darwin":
		if envShell != nil {
			return *envShell
		}
		return UNIX_DEFAULT_SHELL
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
