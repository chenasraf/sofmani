package utils

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"strings"

	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
	"github.com/samber/lo"
)

// ResolveEnvPaths takes one or more slices of environment variable strings (e.g., "KEY=VALUE"),
// resolves any paths within the values using GetRealPath, and returns a single combined slice.
func ResolveEnvPaths(envs ...[]string) []string {
	out := []string{}
	for _, e := range envs {
		for _, env := range e {
			k, v, ok := strings.Cut(env, "=")
			if !ok {
				continue
			}
			out = append(out, fmt.Sprintf("%s=%s", k, GetRealPath(e, v)))
		}
	}
	return out
}

// CommandEnv returns the environment for a child command: the process environment, with env
// applied on top after resolving its values (see ResolveEnvPaths). Process values are passed on
// as they are: they are already real values, and resolving them again would rewrite any `$` in a
// secret.
func CommandEnv(env []string) []string {
	return append(os.Environ(), ResolveEnvPaths(env)...)
}

// ResolveEnvMap resolves the values of env (see ResolveEnvPaths), for config env that is applied
// to the process.
func ResolveEnvMap(env map[string]string) map[string]string {
	return EnvSliceAsMap(ResolveEnvPaths(EnvMapAsSlice(env)))
}

// CombineEnv merges multiple slices of environment variable strings.
// Later slices will override earlier ones if keys conflict.
func CombineEnv(envs ...*[]string) []string {
	out := []string{}
	for _, env := range envs {
		out = mergeEnvs(env, out)
	}
	return out
}

// CombineEnvMaps merges multiple maps of environment variables.
// Later maps will override earlier ones if keys conflict.
func CombineEnvMaps(envs ...*map[string]string) map[string]string {
	out := map[string]string{}
	for _, env := range envs {
		if env == nil {
			continue
		}
		maps.Copy(out, *env)
	}
	return out
}

// EnvSliceAsMap converts a slice of environment variable strings ("KEY=VALUE") to a map.
func EnvSliceAsMap(env []string) map[string]string {
	out := map[string]string{}
	for _, line := range env {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		out[k] = v
	}
	return out
}

// EnvMapAsSlice converts a map of environment variables to a slice of "KEY=VALUE" strings.
func EnvMapAsSlice(env map[string]string) []string {
	return lo.MapToSlice(env, func(k string, v string) string {
		return fmt.Sprintf("%s=%s", k, v)
	})
}

// mergeEnvs helper function to merge a source slice of env strings into a target map (represented as a slice).
// This is an internal helper for CombineEnv.
func mergeEnvs(source *[]string, target []string) []string {
	tgt := EnvSliceAsMap(target)
	if source == nil {
		source = &[]string{} // Treat nil source as empty
	}
	maps.Copy(tgt, EnvSliceAsMap(*source))
	return EnvMapAsSlice(tgt)
}

// envCommandIgnored are variables the shell itself changes between two `env` calls, which are
// not something the command exported.
var envCommandIgnored = map[string]bool{"_": true, "SHLVL": true, "PWD": true, "OLDPWD": true}

// RunEnvCommand runs command in shell and returns the environment variables it exported: those
// that are new or changed after it ran. The command runs in the shell process itself, so a shell
// function that exports variables works, as does `eval "$(loader)"`. With interactive, the shell
// starts with -i so it reads its rc file first (plugins, functions, aliases); whatever that
// startup exports is not returned. Standard input, output and error are passed through so the
// command can prompt, e.g. to unlock a password manager. shell must be POSIX-compatible (sh,
// bash, zsh, ...).
func RunEnvCommand(env []string, shell string, interactive bool, command string) (map[string]string, error) {
	if platform.GetPlatform() == platform.PlatformWindows {
		return nil, fmt.Errorf("env_command is not supported on Windows")
	}
	// The environment is dumped before and after the command to descriptors 3 and 4, so nothing
	// the command or the rc files print can mix into it.
	dumps := make([]*os.File, 2)
	for idx := range dumps {
		f, err := os.CreateTemp("", "sofmani-env")
		if err != nil {
			return nil, fmt.Errorf("failed to create env_command temp file: %w", err)
		}
		defer func() {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}()
		dumps[idx] = f
	}
	script := "env -0 >&3\n" + command + "\n__sofmani_status=$?\nenv -0 >&4\nexit $__sofmani_status"
	args := []string{"-c", script}
	if interactive {
		args = append([]string{"-i"}, args...)
	}
	logger.Debug("Running env_command with %s (interactive: %t)", shell, interactive)
	cmd := exec.Command(shell, args...)
	cmd.Env = CommandEnv(env)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = dumps
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("env_command failed: %w", err)
	}

	snapshots := make([]map[string]string, 2)
	for idx, f := range dumps {
		contents, err := os.ReadFile(f.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read env_command output: %w", err)
		}
		snapshots[idx] = parseNulEnv(contents)
	}
	if len(snapshots[1]) == 0 {
		return nil, fmt.Errorf("env_command exited the shell before its environment could be read")
	}
	out := map[string]string{}
	for k, v := range snapshots[1] {
		if before, ok := snapshots[0][k]; (!ok || before != v) && !envCommandIgnored[k] {
			out[k] = v
		}
	}
	return out, nil
}

// parseNulEnv parses the output of `env -0`.
func parseNulEnv(contents []byte) map[string]string {
	out := map[string]string{}
	for entry := range strings.SplitSeq(string(contents), "\x00") {
		if k, v, ok := strings.Cut(entry, "="); ok && k != "" {
			out[k] = v
		}
	}
	return out
}

// launchEnv is the process environment as sofmani received it, before any config applied its
// own env on top. Variables reset by ScopeEnv go back to these values.
var launchEnv = func() map[string]string {
	out := map[string]string{}
	for _, line := range os.Environ() {
		if k, v, ok := strings.Cut(line, "="); ok {
			out[k] = v
		}
	}
	return out
}()

// ScopeEnv changes the process environment for the duration of a scope. Every key in reset goes
// back to its value at launch (or is unset if it had none), then every key in set is applied.
// The returned function puts each touched key back to its value from before the call.
//
// The process environment is what child commands inherit and what path expansion reads, so a
// value only disappears from both by leaving the process environment.
func ScopeEnv(set map[string]string, reset []string) (restore func(), err error) {
	type prior struct {
		value  string
		exists bool
	}
	previous := map[string]prior{}
	remember := func(k string) {
		if _, seen := previous[k]; !seen {
			v, ok := os.LookupEnv(k)
			previous[k] = prior{value: v, exists: ok}
		}
	}
	restore = func() {
		for k, p := range previous {
			if p.exists {
				_ = os.Setenv(k, p.value)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}

	for _, k := range reset {
		remember(k)
		if v, ok := launchEnv[k]; ok {
			err = os.Setenv(k, v)
		} else {
			err = os.Unsetenv(k)
		}
		if err != nil {
			restore()
			return func() {}, fmt.Errorf("failed to reset environment variable %s: %w", k, err)
		}
	}
	for k, v := range set {
		remember(k)
		if err = os.Setenv(k, v); err != nil {
			restore()
			return func() {}, fmt.Errorf("failed to set environment variable %s: %w", k, err)
		}
	}
	return restore, nil
}
