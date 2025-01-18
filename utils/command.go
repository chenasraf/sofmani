package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
)

const UNIX_DEFAULT_SHELL string = "bash"

func RunCmdPassThrough(env []string, bin string, args ...string) error {
	logger.Debug("Running command: %s %v", bin, args)
	cmd := exec.Command(bin, args...)
	cmd.Env = ResolveEnvPaths(os.Environ(), cmd.Env, env)
	cmd.Stdin = os.Stdin
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
	logger.Debug("Running command: %s %v", bin, args)
	cmd := exec.Command(bin, args...)
	cmd.Env = ResolveEnvPaths(os.Environ(), cmd.Env, env)
	err := cmd.Run()
	if err != nil {
		return nil, false
	}
	return nil, true
}

func RunCmdGetOutput(env []string, bin string, args ...string) ([]byte, error) {
	logger.Debug("Running command: %s %v", bin, args)
	cmd := exec.Command(bin, args...)
	cmd.Env = ResolveEnvPaths(os.Environ(), cmd.Env, env)
	out, err := cmd.Output()
	return out, err
}

func getShellScript(dir string) string {
	var filename string
	switch platform.GetPlatform() {
	case platform.PlatformWindows:
		filename = "install.bat"
	case platform.PlatformLinux, platform.PlatformMacos:
		filename = "install"
	}
	tmpfile := filepath.Join(dir, filename)
	return tmpfile
}

func getScriptContents(script string, envShell *platform.PlatformMap[string]) (string, error) {
	switch platform.GetPlatform() {
	case platform.PlatformWindows:
		preScript := "@echo off"
		postScript := "exit /b %ERRORLEVEL%"
		return fmt.Sprintf("%s\n%s\n\n%s\n", preScript, script, postScript), nil
	case platform.PlatformLinux, platform.PlatformMacos:
		shell := GetOSShell(envShell)
		preScript := fmt.Sprintf("#!/usr/bin/env %s", shell)
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		script = strings.ReplaceAll(script, "~", home)
		postScript := "exit $?"
		return fmt.Sprintf("%s\n%s\n\n%s\n", preScript, script, postScript), nil
	}
	return "", fmt.Errorf("unsupported OS: %s", platform.GetPlatform())
}

func RunCmdAsFile(env []string, contents string, envShell *platform.PlatformMap[string]) error {
	tmpdir, err := os.MkdirTemp("", "sofmani-*")
	if err != nil {
		return err
	}
	tmpfile := getShellScript(tmpdir)
	commandStr, err := getScriptContents(contents, envShell)
	if err != nil {
		return err
	}
	err = os.WriteFile(tmpfile, []byte(commandStr), 0755)
	defer os.RemoveAll(tmpdir)
	if err != nil {
		return err
	}

	shell := GetOSShell(envShell)
	args := GetOSShellArgs(tmpfile)
	return RunCmdPassThrough(env, shell, args...)
}

func GetShellWhich() string {
	switch platform.GetPlatform() {
	case platform.PlatformWindows:
		return "where"
	case platform.PlatformLinux, platform.PlatformMacos:
		return "which"
	}
	return ""
}

func GetOSShell(envShell *platform.PlatformMap[string]) string {
	switch platform.GetPlatform() {
	case platform.PlatformWindows:
		return "cmd"
	case platform.PlatformLinux, platform.PlatformMacos:
		def := os.Getenv("SHELL")
		if def == "" {
			def = UNIX_DEFAULT_SHELL
		}
		if envShell != nil {
			return envShell.ResolveWithFallback(platform.PlatformMap[string]{Linux: &def, MacOS: &def})
		}
		return def
	}
	return ""
}

func GetOSShellArgs(cmd string) []string {
	switch platform.GetPlatform() {
	case platform.PlatformWindows:
		return []string{"/C", cmd + " & exit %ERRORLEVEL%"}
	case platform.PlatformLinux, platform.PlatformMacos:
		return []string{"-c", cmd + "; exit $?"}
	}
	return []string{}
}
