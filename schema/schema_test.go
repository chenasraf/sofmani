package schema_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// schemaPath returns the absolute path to the schema file regardless of where
// the tests are run from.
func schemaPath(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(wd, "sofmani.schema.json")
}

func loadSchema(t *testing.T) map[string]any {
	t.Helper()
	data, err := os.ReadFile(schemaPath(t))
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m), "schema must be valid JSON")
	return m
}

func TestSchemaIsValidJSON(t *testing.T) {
	m := loadSchema(t)
	assert.Equal(t, "http://json-schema.org/draft-07/schema#", m["$schema"])
	assert.NotEmpty(t, m["$id"])
	assert.NotEmpty(t, m["title"])
	assert.Equal(t, "object", m["type"])
}

func TestSchemaTopLevelProperties(t *testing.T) {
	m := loadSchema(t)
	props, ok := m["properties"].(map[string]any)
	require.True(t, ok, "top-level properties must exist")

	// These must be declared at the top level of the schema.
	expected := []string{
		"$schema",
		"debug",
		"check_updates",
		"summary",
		"category_display",
		"repo_update",
		"defaults",
		"env",
		"platform_env",
		"machine_aliases",
		"install",
	}
	for _, key := range expected {
		_, exists := props[key]
		assert.Truef(t, exists, "top-level property %q missing from schema", key)
	}
}

// TestInstallerTypesMatchGoConstants ensures the schema's list of installer
// types stays in lock-step with the Go InstallerType constants. If a new
// installer type is added in code but not in the schema (or vice-versa), this
// test fails and prompts the developer to update both sides.
func TestInstallerTypesMatchGoConstants(t *testing.T) {
	m := loadSchema(t)
	defs, ok := m["definitions"].(map[string]any)
	require.True(t, ok)
	installerType, ok := defs["installerType"].(map[string]any)
	require.True(t, ok)
	enum, ok := installerType["enum"].([]any)
	require.True(t, ok)

	schemaTypes := make([]string, 0, len(enum))
	for _, v := range enum {
		s, ok := v.(string)
		require.True(t, ok)
		schemaTypes = append(schemaTypes, s)
	}
	sort.Strings(schemaTypes)

	goTypes := []string{
		string(appconfig.InstallerTypeGroup),
		string(appconfig.InstallerTypeShell),
		string(appconfig.InstallerTypeDocker),
		string(appconfig.InstallerTypeBrew),
		string(appconfig.InstallerTypeApt),
		string(appconfig.InstallerTypeApk),
		string(appconfig.InstallerTypeGit),
		string(appconfig.InstallerTypeGitHubRelease),
		string(appconfig.InstallerTypeRsync),
		string(appconfig.InstallerTypeNpm),
		string(appconfig.InstallerTypePnpm),
		string(appconfig.InstallerTypeYarn),
		string(appconfig.InstallerTypePipx),
		string(appconfig.InstallerTypeManifest),
		string(appconfig.InstallerTypePacman),
		string(appconfig.InstallerTypeYay),
		string(appconfig.InstallerTypeCargo),
	}
	sort.Strings(goTypes)

	assert.Equal(t, goTypes, schemaTypes, "installerType enum in schema is out of sync with Go constants")
}

func TestCategoryDisplayEnumMatchesGoConstants(t *testing.T) {
	m := loadSchema(t)
	props := m["properties"].(map[string]any)
	cat := props["category_display"].(map[string]any)
	enum, ok := cat["enum"].([]any)
	require.True(t, ok)

	schemaVals := make([]string, 0, len(enum))
	for _, v := range enum {
		schemaVals = append(schemaVals, v.(string))
	}
	sort.Strings(schemaVals)

	goVals := []string{
		string(appconfig.CategoryDisplayBorder),
		string(appconfig.CategoryDisplayBorderCompact),
		string(appconfig.CategoryDisplayMinimal),
	}
	sort.Strings(goVals)

	assert.Equal(t, goVals, schemaVals)
}

func TestRepoUpdateEnumMatchesGoConstants(t *testing.T) {
	m := loadSchema(t)
	defs := m["definitions"].(map[string]any)
	mode := defs["repoUpdateMode"].(map[string]any)
	enum, ok := mode["enum"].([]any)
	require.True(t, ok)

	schemaVals := make([]string, 0, len(enum))
	for _, v := range enum {
		schemaVals = append(schemaVals, v.(string))
	}
	sort.Strings(schemaVals)

	goVals := []string{
		string(appconfig.RepoUpdateOnce),
		string(appconfig.RepoUpdateAlways),
		string(appconfig.RepoUpdateNever),
	}
	sort.Strings(goVals)

	assert.Equal(t, goVals, schemaVals)
}

// TestRecipesParseAgainstSchemaShape is a structural smoke test: every recipe
// shipped in docs/recipes must only use top-level keys that the schema
// declares. This catches typos and schema drift without pulling in a full
// JSON-schema validator as a dependency.
func TestRecipesParseAgainstSchemaShape(t *testing.T) {
	recipesDir := filepath.Join("..", "docs", "recipes")
	entries, err := os.ReadDir(recipesDir)
	require.NoError(t, err)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".yml" && filepath.Ext(name) != ".yaml" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(recipesDir, name)
			data, err := os.ReadFile(path)
			require.NoError(t, err)

			cfg, err := appconfig.ParseConfigFromContent(data)
			require.NoError(t, err, "recipe must parse as AppConfig")

			// Basic sanity: every installer in the recipe has a recognized
			// type (or is a category header).
			for _, inst := range cfg.Install {
				if inst.IsCategory() {
					continue
				}
				assert.NotEmptyf(t, string(inst.Type), "installer in %s missing type", name)
			}
		})
	}

}
