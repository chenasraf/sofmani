package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
)

// UNIX_DEFAULT_SHELL is the default shell used on Unix-like systems if SHELL environment variable is not set.
const UNIX_DEFAULT_SHELL string = "bash"

// RunCmdPassThrough executes a command and passes through its standard input, output, and error streams.
// It also resolves environment variable paths.
func RunCmdPassThrough(env []string, bin string, args ...string) error {
	logger.Debug("Running command: %s %v", bin, args)
	cmd := exec.Command(bin, args...)
	cmd.Env = ResolveEnvPaths(os.Environ(), cmd.Env, env)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunCmdPassThroughChained executes a series of commands sequentially, passing through streams.
// If any command fails, the chain stops and an error is returned.
func RunCmdPassThroughChained(env []string, commands [][]string) error {
	for _, c := range commands {
		err := RunCmdPassThrough(env, c[0], c[1:]...)
		if err != nil {
			return err
		}
	}
	return nil
}

// RunCmdGetSuccess executes a command and returns true if it succeeds (exit code 0).
// Standard input, output, and error are not passed through.
func RunCmdGetSuccess(env []string, bin string, args ...string) (bool, error) {
	logger.Debug("Running command: %s %v", bin, args)
	cmd := exec.Command(bin, args...)
	cmd.Env = ResolveEnvPaths(os.Environ(), cmd.Env, env)
	err := cmd.Run()
	if err != nil {
		return false, nil // Error means command failed, not an error in execution of this function
	}
	return true, nil
}

// RunCmdGetSuccessPassThrough executes a command, passes through streams, and returns true if it succeeds.
func RunCmdGetSuccessPassThrough(env []string, bin string, args ...string) (bool, error) {
	logger.Debug("Running command: %s %v", bin, args)
	cmd := exec.Command(bin, args...)
	cmd.Env = ResolveEnvPaths(os.Environ(), cmd.Env, env)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return false, nil
	}
	return true, nil
}

// RunCmdGetOutput executes a command and returns its standard output.
func RunCmdGetOutput(env []string, bin string, args ...string) ([]byte, error) {
	logger.Debug("Running command: %s %v", bin, args)
	cmd := exec.Command(bin, args...)
	cmd.Env = ResolveEnvPaths(os.Environ(), cmd.Env, env)
	out, err := cmd.Output()
	return out, err
}

// getShellScript returns the appropriate shell script filename based on the OS.
func getShellScript(dir string) string {
	var filename string
	switch platform.GetPlatform() {
	case platform.PlatformWindows:
		filename = "install.bat"
	case platform.PlatformLinux, platform.PlatformMacos:
		filename = "install" // Typically no extension needed for Unix shell scripts
	}
	tmpfile := filepath.Join(dir, filename)
	return tmpfile
}

// getScriptContents prepares the script content with OS-specific shebangs and exit commands.
func getScriptContents(script string, envShell *platform.PlatformMap[string]) (string, error) {
	switch platform.GetPlatform() {
	case platform.PlatformWindows:
		preScript := "@echo off"
		postScript := "exit /b %ERRORLEVEL%" // Ensures the script's exit code is propagated
		return fmt.Sprintf("%s\n%s\n\n%s\n", preScript, script, postScript), nil
	case platform.PlatformLinux, platform.PlatformMacos:
		shell := GetOSShell(envShell)
		preScript := fmt.Sprintf("#!/usr/bin/env %s", shell)
		home, err := os.UserHomeDir() // For resolving ~ in paths
		if err != nil {
			return "", err
		}
		script = strings.ReplaceAll(script, "~", home)
		postScript := "exit $?" // Ensures the script's exit code is propagated
		return fmt.Sprintf("%s\n%s\n\n%s\n", preScript, script, postScript), nil
	}
	return "", fmt.Errorf("unsupported OS: %s", platform.GetPlatform())
}

// RunCmdAsFile writes the given contents to a temporary shell script and executes it.
// This is useful for running multi-line commands or scripts.
func RunCmdAsFile(env []string, contents string, envShell *platform.PlatformMap[string]) error {
	tmpdir, err := os.MkdirTemp("", "sofmani-*") // Create a temporary directory
	if err != nil {
		return err
	}
	tmpfile := getShellScript(tmpdir) // Get OS-specific script name
	commandStr, err := getScriptContents(contents, envShell)
	if err != nil {
		return err
	}
	err = os.WriteFile(tmpfile, []byte(commandStr), 0755) // Make executable
	defer os.RemoveAll(tmpdir)                            // Clean up the temporary directory
	if err != nil {
		return err
	}

	shell := GetOSShell(envShell)
	args := GetOSShellArgs(tmpfile) // Get OS-specific arguments to run the script
	logger.Debug("Running command as file: %s", contents)
	return RunCmdPassThrough(env, shell, args...)
}

// GetShellWhich returns the command used to find the path of an executable (e.g., "which" or "where").
func GetShellWhich() string {
	switch platform.GetPlatform() {
	case platform.PlatformWindows:
		return "where"
	case platform.PlatformLinux, platform.PlatformMacos:
		return "which"
	}
	return ""
}

// GetOSShell returns the appropriate shell for the current operating system.
// It considers the SHELL environment variable on Unix-like systems and allows overrides via envShell.
func GetOSShell(envShell *platform.PlatformMap[string]) string {
	switch platform.GetPlatform() {
	case platform.PlatformWindows:
		return "cmd"
	case platform.PlatformLinux, platform.PlatformMacos:
		def := os.Getenv("SHELL") // Use user's preferred shell if set
		if def == "" {
			def = UNIX_DEFAULT_SHELL // Fallback to bash
		}
		if envShell != nil {
			// Allow platform-specific override from installer config
			return envShell.ResolveWithFallback(platform.PlatformMap[string]{Linux: &def, MacOS: &def})
		}
		return def
	}
	return ""
}

// GetOSShellArgs returns the appropriate shell arguments for executing a command string.
func GetOSShellArgs(cmd string) []string {
	switch platform.GetPlatform() {
	case platform.PlatformWindows:
		// cmd /C "command & exit %ERRORLEVEL%" ensures the exit code is propagated.
		return []string{"/C", cmd + " & exit %ERRORLEVEL%"}
	case platform.PlatformLinux, platform.PlatformMacos:
		// shell -c "command; exit $?" ensures the exit code is propagated.
		return []string{"-c", cmd + "; exit $?"}
	}
	return []string{}
}
