package utils

import (
	"fmt"
	"strings"
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
		for k, v := range *env {
			out[k] = v
		}
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
	out := []string{}
	for k, v := range env {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}

// mergeEnvs helper function to merge a source slice of env strings into a target map (represented as a slice).
// This is an internal helper for CombineEnv.
func mergeEnvs(source *[]string, target []string) []string {
	tgt := EnvSliceAsMap(target)
	if source == nil {
		source = &[]string{} // Treat nil source as empty
	}
	for k, v := range EnvSliceAsMap(*source) {
		tgt[k] = v // Override or add keys from source
	}
	return EnvMapAsSlice(tgt)
}
