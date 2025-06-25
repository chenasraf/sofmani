package installer

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
	"github.com/stretchr/testify/assert"
)

type MockInstaller struct {
	data             *appconfig.InstallerData
	isInstalled      bool
	needsUpdate      bool
	installError     error
	updateError      error
	checkInstall     error
	checkUpdate      error
	validationErrors []ValidationError
}

func (m *MockInstaller) GetData() *appconfig.InstallerData {
	return m.data
}

func (m *MockInstaller) CheckIsInstalled() (bool, error) {
	return m.isInstalled, m.checkInstall
}

func (m *MockInstaller) CheckNeedsUpdate() (bool, error) {
	return m.needsUpdate, m.checkUpdate
}

func (m *MockInstaller) Install() error {
	return m.installError
}

func (m *MockInstaller) Update() error {
	return m.updateError
}

func (m *MockInstaller) Validate() []ValidationError {
	return m.validationErrors
}

func TestGetInstaller(t *testing.T) {
	config := &appconfig.AppConfig{}
	logger.InitLogger(false)
	installer := &appconfig.InstallerData{Type: appconfig.InstallerTypeBrew}
	inst, err := GetInstaller(config, installer)
	assert.NoError(t, err)
	assert.NotNil(t, inst)
}

func TestInstallerWithDefaults(t *testing.T) {
	opts := map[string]any{"key": "value"}
	defaults := &appconfig.AppConfigDefaults{
		Type: &map[appconfig.InstallerType]appconfig.InstallerData{
			appconfig.InstallerTypeBrew: {Opts: &opts},
		},
	}
	installer := &appconfig.InstallerData{Type: appconfig.InstallerTypeBrew, Opts: &map[string]any{}}
	result := InstallerWithDefaults(installer, appconfig.InstallerTypeBrew, defaults)
	assert.Equal(t, "value", (*result.Opts)["key"])
}

func TestRunInstaller(t *testing.T) {
	config := &appconfig.AppConfig{}
	mockInstaller := &MockInstaller{
		data:        &appconfig.InstallerData{Name: strPtr("test"), Type: appconfig.InstallerTypeBrew},
		isInstalled: false,
	}
	err := RunInstaller(config, mockInstaller)
	assert.NoError(t, err)
}

func TestAptValidation(t *testing.T) {
	logger.InitLogger(false)
	aptInstaller := &AptInstaller{
		InstallerBase: InstallerBase{
			Data: &appconfig.InstallerData{
				Name: strPtr("test-apt"),
				Type: appconfig.InstallerTypeApt,
			},
		},
	}
	errors := aptInstaller.Validate()
	assert.Empty(t, errors)
}

func newTestBrewInstaller(data *appconfig.InstallerData) *BrewInstaller {
	return &BrewInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Info: data,
	}
}

func simulateBrewNeedsUpdateFilter(input string) (string, error) {
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("echo '%s' %s", input, PipedInputNeedsUpdateCommand),
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestBrewValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid: No tap specified (tap is optional)
	emptyData := &appconfig.InstallerData{
		Name: strPtr("test-brew-valid"),
		Type: appconfig.InstallerTypeBrew,
	}
	assert.Empty(t, newTestBrewInstaller(emptyData).Validate())

	// 🟢 Valid: Well-formed tap (contains slash, sufficient length)
	validData := &appconfig.InstallerData{
		Name: strPtr("test-brew-valid-tap"),
		Type: appconfig.InstallerTypeBrew,
		Opts: &map[string]any{"tap": "valid/tap"},
	}
	assert.Empty(t, newTestBrewInstaller(validData).Validate())

	// 🔴 Invalid: Tap is present but malformed (missing slash or too short)
	invalidData := &appconfig.InstallerData{
		Name: strPtr("test-brew-invalid-tap"),
		Type: appconfig.InstallerTypeBrew,
		Opts: &map[string]any{"tap": "invalid-tap"},
	}
	assert.NotEmpty(t, newTestBrewInstaller(invalidData).Validate())
}

func TestBrewNeedsUpdateFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "filters empty JSON",
			input: `{
  "formulae": [],
  "casks": []
}`,
			expected: "",
		},
		{
			name: "keeps non-empty JSON",
			input: `{
  "formulae": [{ "name": "foo", "current_version": "1.0" }],
  "casks": []
}`,
			expected: `{
  "formulae": [{ "name": "foo", "current_version": "1.0" }],
  "casks": []
}`,
		},
		{
			name: "keeps extra output lines",
			input: `Warning: You have unlinked kegs
{
  "formulae": [],
  "casks": []
}`,
			expected: "Warning: You have unlinked kegs",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := simulateBrewNeedsUpdateFilter(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := strings.TrimSpace(output)
			want := strings.TrimSpace(tc.expected)
			if got != want {
				t.Errorf("unexpected output\nGot:\n%q\nWant:\n%q", got, want)
			}
		})
	}
}

func newTestGitInstaller(data *appconfig.InstallerData) *GitInstaller {
	return &GitInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Info: data,
	}
}

func TestGitValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid: Both destination and ref are present
	validData := &appconfig.InstallerData{
		Name: strPtr("test-git-valid"),
		Type: appconfig.InstallerTypeGit,
		Opts: &map[string]any{
			"destination": "/some/path",
			"ref":         "main",
		},
	}
	errors := newTestGitInstaller(validData).Validate()
	assert.Empty(t, errors)

	// 🔴 Invalid: Missing ref
	missingRefData := &appconfig.InstallerData{
		Name: strPtr("test-git-missing-ref"),
		Type: appconfig.InstallerTypeGit,
		Opts: &map[string]any{
			"destination": "/some/path",
		},
	}
	errors = newTestGitInstaller(missingRefData).Validate()
	assert.Empty(t, errors)

	// 🔴 Invalid: Missing destination
	missingDestData := &appconfig.InstallerData{
		Name: strPtr("test-git-missing-destination"),
		Type: appconfig.InstallerTypeGit,
		Opts: &map[string]any{
			"ref": "main",
		},
	}
	errors = newTestGitInstaller(missingDestData).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "destination", errors[0].FieldName)

	// 🔴 Invalid: Missing both destination and ref
	missingBothData := &appconfig.InstallerData{
		Name: strPtr("test-git-missing-both"),
		Type: appconfig.InstallerTypeGit,
		Opts: &map[string]any{},
	}
	errors = newTestGitInstaller(missingBothData).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "destination", errors[0].FieldName)
}

func newTestGitHubReleaseInstaller(data *appconfig.InstallerData) *GitHubReleaseInstaller {
	return &GitHubReleaseInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Info: data,
	}
}

func TestGitHubReleaseValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid
	validData := &appconfig.InstallerData{
		Name: strPtr("ghr-valid"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"repository":        "owner/repo",
			"destination":       "/some/path",
			"download_filename": "file.tar.gz", // valid string
			"strategy":          "tar",
		},
	}
	assert.Empty(t, newTestGitHubReleaseInstaller(validData).Validate())

	// 🔴 Missing repository
	missingRepo := &appconfig.InstallerData{
		Name: strPtr("ghr-missing-repo"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"destination":       "/some/path",
			"download_filename": "file.tar.gz",
		},
	}
	errors := newTestGitHubReleaseInstaller(missingRepo).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "repository", errors[0].FieldName)

	// 🔴 Missing download_filename
	missingDownloadFilename := &appconfig.InstallerData{
		Name: strPtr("ghr-missing-download"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"repository":  "owner/repo",
			"destination": "/some/path",
		},
	}
	errors = newTestGitHubReleaseInstaller(missingDownloadFilename).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "download_filename", errors[0].FieldName)

	// 🔴 Empty per-platform download_filename
	emptyPlatformFilename := &appconfig.InstallerData{
		Name: strPtr("ghr-empty-platform-filename"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"repository":  "owner/repo",
			"destination": "/some/path",
			"download_filename": map[string]*string{
				string(platform.GetPlatform()): strPtr(""),
			},
		},
	}
	errors = newTestGitHubReleaseInstaller(emptyPlatformFilename).Validate()
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0].FieldName, "download_filename")

	// 🔴 Invalid strategy
	invalidStrategy := &appconfig.InstallerData{
		Name: strPtr("ghr-invalid-strategy"),
		Type: appconfig.InstallerTypeGitHubRelease,
		Opts: &map[string]any{
			"repository":        "owner/repo",
			"destination":       "/some/path",
			"download_filename": "file.tar.gz",
			"strategy":          "exe", // invalid
		},
	}
	errors = newTestGitHubReleaseInstaller(invalidStrategy).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "strategy", errors[0].FieldName)
}

func newTestGroupInstaller(data *appconfig.InstallerData) *GroupInstaller {
	return &GroupInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config: nil,
		Data:   data,
	}
}

func TestGroupValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid: one sub-installer
	validStep := appconfig.InstallerData{
		Name: strPtr("child-installer"),
		Type: appconfig.InstallerTypeBrew,
	}
	validData := &appconfig.InstallerData{
		Name:  strPtr("group-valid"),
		Type:  appconfig.InstallerTypeGroup,
		Steps: &[]appconfig.InstallerData{validStep},
	}
	assert.Empty(t, newTestGroupInstaller(validData).Validate())

	// 🔴 Invalid: empty steps slice
	emptySteps := &appconfig.InstallerData{
		Name:  strPtr("group-empty"),
		Type:  appconfig.InstallerTypeGroup,
		Steps: &[]appconfig.InstallerData{},
	}
	errors := newTestGroupInstaller(emptySteps).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "steps", errors[0].FieldName)

	// 🔴 Invalid: nil steps
	nilSteps := &appconfig.InstallerData{
		Name:  strPtr("group-nil"),
		Type:  appconfig.InstallerTypeGroup,
		Steps: nil,
	}
	errors = newTestGroupInstaller(nilSteps).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "steps", errors[0].FieldName)
}

func newTestManifestInstaller(data *appconfig.InstallerData) *ManifestInstaller {
	return &ManifestInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config: nil,
		Info:   data,
	}
}

func TestManifestValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid
	validData := &appconfig.InstallerData{
		Name: strPtr("manifest-valid"),
		Type: appconfig.InstallerTypeManifest,
		Opts: &map[string]any{
			"source": "https://example.com/repo.git",
			"path":   "manifests/installer.yml",
			"ref":    "main",
		},
	}
	assert.Empty(t, newTestManifestInstaller(validData).Validate())

	// 🔴 Missing source
	missingSource := &appconfig.InstallerData{
		Name: strPtr("manifest-missing-source"),
		Type: appconfig.InstallerTypeManifest,
		Opts: &map[string]any{
			"path": "some/path",
		},
	}
	errors := newTestManifestInstaller(missingSource).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "source", errors[0].FieldName)

	// 🔴 Missing path
	missingPath := &appconfig.InstallerData{
		Name: strPtr("manifest-missing-path"),
		Type: appconfig.InstallerTypeManifest,
		Opts: &map[string]any{
			"source": "https://example.com/repo.git",
		},
	}
	errors = newTestManifestInstaller(missingPath).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "path", errors[0].FieldName)

	// 🔴 Empty ref (not nil, just empty)
	emptyRef := &appconfig.InstallerData{
		Name: strPtr("manifest-empty-ref"),
		Type: appconfig.InstallerTypeManifest,
		Opts: &map[string]any{
			"source": "https://example.com/repo.git",
			"path":   "install.yml",
			"ref":    "",
		},
	}
	errors = newTestManifestInstaller(emptyRef).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "ref", errors[0].FieldName)
}

func newTestNpmInstaller(data *appconfig.InstallerData) *NpmInstaller {
	return &NpmInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config:         nil,
		PackageManager: PackageManagerNpm,
		Info:           data,
	}
}

func TestNpmValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid npm installer
	validData := &appconfig.InstallerData{
		Name: strPtr("some-npm-package"),
		Type: appconfig.InstallerTypeNpm,
	}
	assert.Empty(t, newTestNpmInstaller(validData).Validate())

	// 🔴 Edge case: nil name (will panic or fail in BaseValidate if implemented to check it)
	nilNameData := &appconfig.InstallerData{
		Name: nil,
		Type: appconfig.InstallerTypeNpm,
	}
	errors := newTestNpmInstaller(nilNameData).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "name", errors[0].FieldName)
}

func newTestPipxInstaller(data *appconfig.InstallerData) *PipxInstaller {
	return &PipxInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config: nil,
		Info:   data,
	}
}

func TestPipxValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid pipx installer
	validData := &appconfig.InstallerData{
		Name: strPtr("some-pipx-package"),
		Type: appconfig.InstallerTypePipx,
	}
	assert.Empty(t, newTestPipxInstaller(validData).Validate())

	// 🔴 Optional: test nil name if BaseValidate handles it
	nilNameData := &appconfig.InstallerData{
		Name: nil,
		Type: appconfig.InstallerTypePipx,
	}
	errors := newTestPipxInstaller(nilNameData).Validate()
	// Uncomment if BaseValidate checks for nil
	assert.Len(t, errors, 1)
	assert.Equal(t, "name", errors[0].FieldName)
}

func newTestRsyncInstaller(data *appconfig.InstallerData) *RsyncInstaller {
	return &RsyncInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config: nil,
		Info:   data,
	}
}

func TestRsyncValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid rsync config
	validData := &appconfig.InstallerData{
		Name: strPtr("rsync-valid"),
		Type: appconfig.InstallerTypeRsync,
		Opts: &map[string]any{
			"source":      "/path/from",
			"destination": "/path/to",
			"flags":       "-avz",
		},
	}
	assert.Empty(t, newTestRsyncInstaller(validData).Validate())

	// 🔴 Missing source
	missingSource := &appconfig.InstallerData{
		Name: strPtr("rsync-missing-source"),
		Type: appconfig.InstallerTypeRsync,
		Opts: &map[string]any{
			"destination": "/path/to",
		},
	}
	errors := newTestRsyncInstaller(missingSource).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "source", errors[0].FieldName)

	// 🔴 Missing destination
	missingDest := &appconfig.InstallerData{
		Name: strPtr("rsync-missing-destination"),
		Type: appconfig.InstallerTypeRsync,
		Opts: &map[string]any{
			"source": "/path/from",
		},
	}
	errors = newTestRsyncInstaller(missingDest).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "destination", errors[0].FieldName)

	// 🔴 Empty flags string
	emptyFlags := &appconfig.InstallerData{
		Name: strPtr("rsync-empty-flags"),
		Type: appconfig.InstallerTypeRsync,
		Opts: &map[string]any{
			"source":      "/path/from",
			"destination": "/path/to",
			"flags":       "",
		},
	}
	errors = newTestRsyncInstaller(emptyFlags).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "flags", errors[0].FieldName)
}

func newTestShellInstaller(data *appconfig.InstallerData) *ShellInstaller {
	return &ShellInstaller{
		InstallerBase: InstallerBase{Data: data},
		Config:        nil,
		Info:          data,
	}
}

func TestShellValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid shell config
	validData := &appconfig.InstallerData{
		Name: strPtr("shell-valid"),
		Type: appconfig.InstallerTypeShell,
		Opts: &map[string]any{
			"command":        "echo install",
			"update_command": "echo update",
		},
	}
	assert.Empty(t, newTestShellInstaller(validData).Validate())

	// 🔴 Missing command
	missingCommand := &appconfig.InstallerData{
		Name: strPtr("shell-missing-command"),
		Type: appconfig.InstallerTypeShell,
		Opts: &map[string]any{
			"update_command": "echo update",
		},
	}
	errors := newTestShellInstaller(missingCommand).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "command", errors[0].FieldName)

	// 🔴 Missing update_command
	missingUpdate := &appconfig.InstallerData{
		Name: strPtr("shell-missing-update"),
		Type: appconfig.InstallerTypeShell,
		Opts: &map[string]any{
			"command": "echo install",
		},
	}
	errors = newTestShellInstaller(missingUpdate).Validate()
	assert.Empty(t, errors)

	// 🔴 Missing both
	missingBoth := &appconfig.InstallerData{
		Name: strPtr("shell-missing-both"),
		Type: appconfig.InstallerTypeShell,
		Opts: &map[string]any{},
	}
	errors = newTestShellInstaller(missingBoth).Validate()
	assert.Len(t, errors, 1)
	assert.Equal(t, "command", errors[0].FieldName)
}

func strPtr(s string) *string {
	return &s
}
