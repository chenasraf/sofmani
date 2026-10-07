package utils

import (
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/samber/lo"
)

// ResolveEnvPaths takes one or more slices of environment variable strings (e.g., "KEY=VALUE"),
// resolves any paths within the values using GetRealPath, and returns a single combined slice.
func ResolveEnvPaths(envs ...[]string) []string {
	out := []string{}
	for _, e := range envs {
		for _, env := range e {
			vals := strings.Split(env, "=")
			if len(vals) != 2 {
				continue
			}
			out = append(out, fmt.Sprintf("%s=%s", vals[0], GetRealPath(e, vals[1])))
		}
	}
	return out
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
		vals := strings.Split(line, "=")
		if len(vals) != 2 {
			continue
		}
		k := vals[0]
		v := vals[1]
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
