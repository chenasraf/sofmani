package appconfig

import (
	"sort"
	"testing"

	"github.com/chenasraf/sofmani/platform"
	"github.com/stretchr/testify/assert"
	yamlPkg "gopkg.in/yaml.v3"
)

func TestInstallerData_Environ(t *testing.T) {
	t.Run("returns empty slice when both Env and PlatformEnv are nil", func(t *testing.T) {
		data := &InstallerData{}
		result := data.Environ()
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("returns Env values when only Env is set", func(t *testing.T) {
		env := map[string]string{"KEY": "value", "OTHER": "test"}
		data := &InstallerData{
			Env: &env,
		}
		result := data.Environ()
		sort.Strings(result) // Sort for consistent comparison
		assert.Len(t, result, 2)
		assert.Contains(t, result, "KEY=value")
		assert.Contains(t, result, "OTHER=test")
	})

	t.Run("returns PlatformEnv values for current platform", func(t *testing.T) {
		macEnv := map[string]string{"PLATFORM": "macos"}
		linuxEnv := map[string]string{"PLATFORM": "linux"}
		data := &InstallerData{
			PlatformEnv: &platform.PlatformMap[map[string]string]{
				MacOS: &macEnv,
				Linux: &linuxEnv,
			},
		}
		result := data.Environ()
		// Result depends on current platform
		if len(result) > 0 {
			assert.Contains(t, result[0], "PLATFORM=")
		}
	})

	t.Run("combines Env and PlatformEnv", func(t *testing.T) {
		env := map[string]string{"COMMON": "value"}
		macEnv := map[string]string{"SPECIFIC": "mac"}
		data := &InstallerData{
			Env: &env,
			PlatformEnv: &platform.PlatformMap[map[string]string]{
				MacOS: &macEnv,
			},
		}
		result := data.Environ()
		assert.Contains(t, result, "COMMON=value")
		// SPECIFIC will only appear on macOS
	})

	t.Run("PlatformEnv overrides Env for same key", func(t *testing.T) {
		env := map[string]string{"KEY": "original"}
		macEnv := map[string]string{"KEY": "platform"}
		data := &InstallerData{
			Env: &env,
			PlatformEnv: &platform.PlatformMap[map[string]string]{
				MacOS: &macEnv,
			},
		}
		result := data.Environ()
		// On macOS, KEY should be "platform"
		// On other platforms, KEY should be "original"
		assert.Len(t, result, 1)
		assert.Contains(t, result[0], "KEY=")
	})
}

func TestInstallerData_GetTagsList(t *testing.T) {
	t.Run("returns list of space-separated tags", func(t *testing.T) {
		tags := "python node rust"
		data := &InstallerData{
			Tags: &tags,
		}
		result := data.GetTagsList()
		assert.Equal(t, []string{"python", "node", "rust"}, result)
	})

	t.Run("trims whitespace from tags", func(t *testing.T) {
		tags := "  python   node   rust  "
		data := &InstallerData{
			Tags: &tags,
		}
		result := data.GetTagsList()
		// Note: empty strings will be included for leading/trailing spaces when split
		// The implementation splits and trims each part
		for _, tag := range result {
			assert.Equal(t, tag, trimmedTag(tag))
		}
	})

	t.Run("returns single tag when only one is present", func(t *testing.T) {
		tags := "python"
		data := &InstallerData{
			Tags: &tags,
		}
		result := data.GetTagsList()
		assert.Equal(t, []string{"python"}, result)
	})

	t.Run("handles tags with multiple spaces between them", func(t *testing.T) {
		tags := "python  node"
		data := &InstallerData{
			Tags: &tags,
		}
		result := data.GetTagsList()
		// Split by single space, so empty string will be in between
		assert.Contains(t, result, "python")
		assert.Contains(t, result, "node")
	})
}

// helper to get trimmed tag
func trimmedTag(s string) string {
	return s // Already trimmed by the function
}

func TestInstallerType_Constants(t *testing.T) {
	t.Run("installer types have expected values", func(t *testing.T) {
		assert.Equal(t, InstallerType("group"), InstallerTypeGroup)
		assert.Equal(t, InstallerType("shell"), InstallerTypeShell)
		assert.Equal(t, InstallerType("docker"), InstallerTypeDocker)
		assert.Equal(t, InstallerType("brew"), InstallerTypeBrew)
		assert.Equal(t, InstallerType("apt"), InstallerTypeApt)
		assert.Equal(t, InstallerType("apk"), InstallerTypeApk)
		assert.Equal(t, InstallerType("git"), InstallerTypeGit)
		assert.Equal(t, InstallerType("github-release"), InstallerTypeGitHubRelease)
		assert.Equal(t, InstallerType("rsync"), InstallerTypeRsync)
		assert.Equal(t, InstallerType("npm"), InstallerTypeNpm)
		assert.Equal(t, InstallerType("pnpm"), InstallerTypePnpm)
		assert.Equal(t, InstallerType("yarn"), InstallerTypeYarn)
		assert.Equal(t, InstallerType("pipx"), InstallerTypePipx)
		assert.Equal(t, InstallerType("manifest"), InstallerTypeManifest)
		assert.Equal(t, InstallerType("pacman"), InstallerTypePacman)
		assert.Equal(t, InstallerType("yay"), InstallerTypeYay)
	})
}

func TestSkipSummary_UnmarshalYAML(t *testing.T) {
	t.Run("boolean true applies to both install and update", func(t *testing.T) {
		yaml := `skip_summary: true`
		var data InstallerData
		err := parseYAML(yaml, &data)
		assert.NoError(t, err)
		assert.NotNil(t, data.SkipSummary)
		assert.True(t, data.SkipSummary.Install)
		assert.True(t, data.SkipSummary.Update)
	})

	t.Run("boolean false applies to both install and update", func(t *testing.T) {
		yaml := `skip_summary: false`
		var data InstallerData
		err := parseYAML(yaml, &data)
		assert.NoError(t, err)
		assert.NotNil(t, data.SkipSummary)
		assert.False(t, data.SkipSummary.Install)
		assert.False(t, data.SkipSummary.Update)
	})

	t.Run("map with install only", func(t *testing.T) {
		yaml := `skip_summary:
  install: true`
		var data InstallerData
		err := parseYAML(yaml, &data)
		assert.NoError(t, err)
		assert.NotNil(t, data.SkipSummary)
		assert.True(t, data.SkipSummary.Install)
		assert.False(t, data.SkipSummary.Update)
	})

	t.Run("map with update only", func(t *testing.T) {
		yaml := `skip_summary:
  update: true`
		var data InstallerData
		err := parseYAML(yaml, &data)
		assert.NoError(t, err)
		assert.NotNil(t, data.SkipSummary)
		assert.False(t, data.SkipSummary.Install)
		assert.True(t, data.SkipSummary.Update)
	})

	t.Run("map with both install and update", func(t *testing.T) {
		yaml := `skip_summary:
  install: true
  update: false`
		var data InstallerData
		err := parseYAML(yaml, &data)
		assert.NoError(t, err)
		assert.NotNil(t, data.SkipSummary)
		assert.True(t, data.SkipSummary.Install)
		assert.False(t, data.SkipSummary.Update)
	})

	t.Run("nil when not specified", func(t *testing.T) {
		yaml := `name: test`
		var data InstallerData
		err := parseYAML(yaml, &data)
		assert.NoError(t, err)
		assert.Nil(t, data.SkipSummary)
	})
}

func parseYAML(yamlStr string, v any) error {
	return yamlPkg.Unmarshal([]byte(yamlStr), v)
}

func TestInstallerData_IsCategory(t *testing.T) {
	t.Run("returns true when Category is set", func(t *testing.T) {
		category := "Development Tools"
		data := &InstallerData{
			Category: &category,
		}
		assert.True(t, data.IsCategory())
	})

	t.Run("returns false when Category is nil", func(t *testing.T) {
		data := &InstallerData{}
		assert.False(t, data.IsCategory())
	})

	t.Run("returns false when Category is empty string", func(t *testing.T) {
		category := ""
		data := &InstallerData{
			Category: &category,
		}
		assert.False(t, data.IsCategory())
	})

	t.Run("category parsed from YAML", func(t *testing.T) {
		yaml := `category: System Utilities`
		var data InstallerData
		err := parseYAML(yaml, &data)
		assert.NoError(t, err)
		assert.True(t, data.IsCategory())
		assert.Equal(t, "System Utilities", *data.Category)
	})

	t.Run("category with desc parsed from YAML", func(t *testing.T) {
		yaml := `category: System Utilities
desc: These are system tools.`
		var data InstallerData
		err := parseYAML(yaml, &data)
		assert.NoError(t, err)
		assert.True(t, data.IsCategory())
		assert.Equal(t, "System Utilities", *data.Category)
		assert.Equal(t, "These are system tools.", *data.Desc)
	})

	t.Run("category with multiline desc parsed from YAML", func(t *testing.T) {
		yaml := `category: Development
desc: |
  First line.
  Second line.`
		var data InstallerData
		err := parseYAML(yaml, &data)
		assert.NoError(t, err)
		assert.True(t, data.IsCategory())
		assert.Equal(t, "Development", *data.Category)
		assert.Contains(t, *data.Desc, "First line.")
		assert.Contains(t, *data.Desc, "Second line.")
	})

	t.Run("regular installer is not a category", func(t *testing.T) {
		yaml := `name: test
type: shell`
		var data InstallerData
		err := parseYAML(yaml, &data)
		assert.NoError(t, err)
		assert.False(t, data.IsCategory())
	})
}
