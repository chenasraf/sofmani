package utils

import (
	"testing"

	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger.InitLogger(false)
}

func TestRunCmdGetSuccess(t *testing.T) {
	tests := []struct {
		name           string
		bin            string
		args           []string
		expectedResult bool
	}{
		{
			name:           "successful command",
			bin:            "echo",
			args:           []string{"hello"},
			expectedResult: true,
		},
		{
			name:           "failing command",
			bin:            "false",
			args:           []string{},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RunCmdGetSuccess(nil, tt.bin, tt.args...)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestRunCmdGetOutput(t *testing.T) {
	tests := []struct {
		name           string
		bin            string
		args           []string
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "echo command",
			bin:            "echo",
			args:           []string{"hello"},
			expectedOutput: "hello\n",
			expectError:    false,
		},
		{
			name:           "printf command",
			bin:            "printf",
			args:           []string{"test"},
			expectedOutput: "test",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := RunCmdGetOutput(nil, tt.bin, tt.args...)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, string(output))
			}
		})
	}
}

func TestRunCmdGetOutputWithEnv(t *testing.T) {
	env := []string{"TEST_VAR=hello_world"}
	output, err := RunCmdGetOutput(env, "sh", "-c", "echo $TEST_VAR")
	assert.NoError(t, err)
	assert.Equal(t, "hello_world\n", string(output))
}

func TestRunCmdPassThroughChained(t *testing.T) {
	// Test successful chain
	commands := [][]string{
		{"true"},
		{"true"},
	}
	err := RunCmdPassThroughChained(nil, commands)
	assert.NoError(t, err)

	// Test chain with failure in middle
	commandsWithFailure := [][]string{
		{"true"},
		{"false"},
		{"true"},
	}
	err = RunCmdPassThroughChained(nil, commandsWithFailure)
	assert.Error(t, err)
}

func TestGetShellWhich(t *testing.T) {
	result := GetShellWhich()
	curPlatform := platform.GetPlatform()

	switch curPlatform {
	case platform.PlatformWindows:
		assert.Equal(t, "where", result)
	case platform.PlatformLinux, platform.PlatformMacos:
		assert.Equal(t, "which", result)
	}
}

func TestGetOSShell(t *testing.T) {
	curPlatform := platform.GetPlatform()

	// Test with nil envShell
	result := GetOSShell(nil)

	switch curPlatform {
	case platform.PlatformWindows:
		assert.Equal(t, "cmd", result)
	case platform.PlatformLinux, platform.PlatformMacos:
		// Should return SHELL env var or default to bash
		assert.NotEmpty(t, result)
	}

	// Test with custom envShell override
	if curPlatform != platform.PlatformWindows {
		customShell := "zsh"
		envShell := &platform.PlatformMap[string]{
			MacOS: &customShell,
			Linux: &customShell,
		}
		result = GetOSShell(envShell)
		assert.Equal(t, "zsh", result)
	}
}

func TestGetOSShellArgs(t *testing.T) {
	curPlatform := platform.GetPlatform()
	args := GetOSShellArgs("echo hello")

	switch curPlatform {
	case platform.PlatformWindows:
		assert.Equal(t, []string{"/C", "echo hello & exit %ERRORLEVEL%"}, args)
	case platform.PlatformLinux, platform.PlatformMacos:
		assert.Equal(t, []string{"-c", "echo hello; exit $?"}, args)
	}
}

func TestGetShellScript(t *testing.T) {
	curPlatform := platform.GetPlatform()
	result := getShellScript("/tmp")

	switch curPlatform {
	case platform.PlatformWindows:
		assert.Equal(t, "/tmp/install.bat", result)
	case platform.PlatformLinux, platform.PlatformMacos:
		assert.Equal(t, "/tmp/install", result)
	}
}

func TestGetScriptContents(t *testing.T) {
	curPlatform := platform.GetPlatform()

	if curPlatform == platform.PlatformWindows {
		content, err := getScriptContents("echo hello", nil)
		assert.NoError(t, err)
		assert.Contains(t, content, "@echo off")
		assert.Contains(t, content, "echo hello")
		assert.Contains(t, content, "exit /b %ERRORLEVEL%")
	} else {
		content, err := getScriptContents("echo hello", nil)
		assert.NoError(t, err)
		assert.Contains(t, content, "#!/usr/bin/env")
		assert.Contains(t, content, "echo hello")
		assert.Contains(t, content, "exit $?")
	}
}

func TestRunCmdAsFile(t *testing.T) {
	curPlatform := platform.GetPlatform()

	if curPlatform != platform.PlatformWindows {
		// Test simple script execution
		err := RunCmdAsFile(nil, "exit 0", nil)
		assert.NoError(t, err)

		// Test script with failure
		err = RunCmdAsFile(nil, "exit 1", nil)
		assert.Error(t, err)
	}
}
