package utils

import (
	"fmt"
	"strings"
)

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

func CombineEnv(envs ...*[]string) []string {
	out := []string{}
	for _, env := range envs {
		out = mergeEnvs(env, out)
	}
	return out
}

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

func EnvMapAsSlice(env map[string]string) []string {
	out := []string{}
	for k, v := range env {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}

func mergeEnvs(source *[]string, target []string) []string {
	tgt := EnvSliceAsMap(target)
	if source == nil {
		source = &[]string{}
	}
	for k, v := range EnvSliceAsMap(*source) {
		tgt[k] = v
	}
	return EnvMapAsSlice(tgt)
}
