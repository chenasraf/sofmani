package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveEnvPaths(t *testing.T) {
	envs := [][]string{
		{"PATH=/usr/bin", "HOME=/home/user"},
		{"GOPATH=/go", "GOROOT=/usr/local/go"},
	}
	expected := []string{
		"PATH=/usr/bin",
		"HOME=/home/user",
		"GOPATH=/go",
		"GOROOT=/usr/local/go",
	}
	result := ResolveEnvPaths(envs...)
	assert.ElementsMatch(t, expected, result)
}

func TestCombineEnv(t *testing.T) {
	env1 := &[]string{"KEY1=value1", "KEY2=value2"}
	env2 := &[]string{"KEY2=new_value2", "KEY3=value3"}
	expected := []string{"KEY1=value1", "KEY2=new_value2", "KEY3=value3"}
	result := CombineEnv(env1, env2)
	assert.ElementsMatch(t, expected, result)
}

func TestCombineEnvMaps(t *testing.T) {
	env1 := &map[string]string{"KEY1": "value1", "KEY2": "value2"}
	env2 := &map[string]string{"KEY2": "new_value2", "KEY3": "value3"}
	expected := map[string]string{"KEY1": "value1", "KEY2": "new_value2", "KEY3": "value3"}
	result := CombineEnvMaps(env1, env2)
	assert.Equal(t, expected, result)
}

func TestEnvSliceAsMap(t *testing.T) {
	env := []string{"KEY1=value1", "KEY2=value2"}
	expected := map[string]string{"KEY1": "value1", "KEY2": "value2"}
	result := EnvSliceAsMap(env)
	assert.Equal(t, expected, result)
}

func TestEnvMapAsSlice(t *testing.T) {
	env := map[string]string{"KEY1": "value1", "KEY2": "value2"}
	expected := []string{"KEY1=value1", "KEY2=value2"}
	result := EnvMapAsSlice(env)
	assert.ElementsMatch(t, expected, result)
}

func TestMergeEnvs(t *testing.T) {
	source := &[]string{"KEY1=value1", "KEY2=value2"}
	target := []string{"KEY2=new_value2", "KEY3=value3"}
	expected := []string{"KEY1=value1", "KEY2=value2", "KEY3=value3"}
	result := mergeEnvs(source, target)
	assert.ElementsMatch(t, expected, result)
}
