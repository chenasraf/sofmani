package utils

import (
	"os"
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

func TestScopeEnv(t *testing.T) {
	launchEnv["SOFMANI_SCOPE_LAUNCH"] = "launch"
	t.Cleanup(func() { delete(launchEnv, "SOFMANI_SCOPE_LAUNCH") })
	t.Setenv("SOFMANI_SCOPE_LAUNCH", "parent")
	t.Setenv("SOFMANI_SCOPE_PARENT_ONLY", "secret")
	t.Setenv("SOFMANI_SCOPE_SET", "before")

	restore, err := ScopeEnv(
		map[string]string{"SOFMANI_SCOPE_SET": "child", "SOFMANI_SCOPE_NEW": "new"},
		[]string{"SOFMANI_SCOPE_LAUNCH", "SOFMANI_SCOPE_PARENT_ONLY"},
	)
	assert.NoError(t, err)

	assert.Equal(t, "launch", os.Getenv("SOFMANI_SCOPE_LAUNCH"))
	_, exists := os.LookupEnv("SOFMANI_SCOPE_PARENT_ONLY")
	assert.False(t, exists)
	assert.Equal(t, "child", os.Getenv("SOFMANI_SCOPE_SET"))
	assert.Equal(t, "new", os.Getenv("SOFMANI_SCOPE_NEW"))

	restore()

	assert.Equal(t, "parent", os.Getenv("SOFMANI_SCOPE_LAUNCH"))
	assert.Equal(t, "secret", os.Getenv("SOFMANI_SCOPE_PARENT_ONLY"))
	assert.Equal(t, "before", os.Getenv("SOFMANI_SCOPE_SET"))
	_, exists = os.LookupEnv("SOFMANI_SCOPE_NEW")
	assert.False(t, exists)
}
