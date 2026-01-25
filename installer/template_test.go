package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
	"github.com/stretchr/testify/assert"
)

func TestNewTemplateVars(t *testing.T) {
	logger.InitLogger(false)

	vars := NewTemplateVars("v1.2.3")
	assert.Equal(t, "v1.2.3", vars.Tag)
	assert.Equal(t, "1.2.3", vars.Version)
	assert.NotEmpty(t, vars.Arch)
	assert.NotEmpty(t, vars.ArchAlias)
	assert.NotEmpty(t, vars.ArchGnu)
	assert.NotEmpty(t, vars.OS)
}

func TestNewTemplateVarsWithoutVPrefix(t *testing.T) {
	logger.InitLogger(false)

	vars := NewTemplateVars("1.2.3")
	assert.Equal(t, "1.2.3", vars.Tag)
	assert.Equal(t, "1.2.3", vars.Version)
}

func TestApplyTemplateGoSyntax(t *testing.T) {
	logger.InitLogger(false)

	// Set predictable values for testing
	platform.SetOS("darwin")
	platform.SetArch("arm64")
	defer func() {
		platform.SetOS("darwin")
		platform.SetArch("arm64")
	}()

	vars := NewTemplateVars("v2.0.0")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Tag variable",
			input:    "app_{{ .Tag }}.tar.gz",
			expected: "app_v2.0.0.tar.gz",
		},
		{
			name:     "Version variable",
			input:    "app_{{ .Version }}.tar.gz",
			expected: "app_2.0.0.tar.gz",
		},
		{
			name:     "Arch variable",
			input:    "app_{{ .Arch }}.tar.gz",
			expected: "app_arm64.tar.gz",
		},
		{
			name:     "ArchAlias variable",
			input:    "app_{{ .ArchAlias }}.tar.gz",
			expected: "app_arm64.tar.gz",
		},
		{
			name:     "OS variable",
			input:    "app_{{ .OS }}.tar.gz",
			expected: "app_macos.tar.gz",
		},
		{
			name:     "ArchGnu variable",
			input:    "app_{{ .ArchGnu }}.tar.gz",
			expected: "app_aarch64.tar.gz",
		},
		{
			name:     "Multiple variables",
			input:    "app_{{ .Version }}_{{ .OS }}_{{ .ArchAlias }}.tar.gz",
			expected: "app_2.0.0_macos_arm64.tar.gz",
		},
		{
			name:     "No variables",
			input:    "app_static.tar.gz",
			expected: "app_static.tar.gz",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ApplyTemplate(tc.input, vars, "test-installer")
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestApplyTemplateLegacySyntax(t *testing.T) {
	logger.InitLogger(false)

	// Set predictable values for testing
	platform.SetOS("linux")
	platform.SetArch("amd64")
	defer func() {
		platform.SetOS("darwin")
		platform.SetArch("arm64")
	}()

	vars := NewTemplateVars("v3.1.4")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Legacy tag token",
			input:    "app_{tag}.tar.gz",
			expected: "app_v3.1.4.tar.gz",
		},
		{
			name:     "Legacy version token",
			input:    "app_{version}.tar.gz",
			expected: "app_3.1.4.tar.gz",
		},
		{
			name:     "Legacy arch token",
			input:    "app_{arch}.tar.gz",
			expected: "app_amd64.tar.gz",
		},
		{
			name:     "Legacy arch_alias token",
			input:    "app_{arch_alias}.tar.gz",
			expected: "app_x86_64.tar.gz",
		},
		{
			name:     "Legacy os token",
			input:    "app_{os}.tar.gz",
			expected: "app_linux.tar.gz",
		},
		{
			name:     "Legacy arch_gnu token",
			input:    "app_{arch_gnu}.tar.gz",
			expected: "app_x86_64.tar.gz",
		},
		{
			name:     "Multiple legacy tokens",
			input:    "app_{version}_{os}_{arch_alias}.tar.gz",
			expected: "app_3.1.4_linux_x86_64.tar.gz",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ApplyTemplate(tc.input, vars, "test-installer")
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestApplyTemplateMixedSyntax(t *testing.T) {
	logger.InitLogger(false)

	// Set predictable values for testing
	platform.SetOS("darwin")
	platform.SetArch("arm64")
	defer func() {
		platform.SetOS("darwin")
		platform.SetArch("arm64")
	}()

	vars := NewTemplateVars("v1.0.0")

	// Mixed syntax should work - legacy tokens are replaced first, then Go template
	result, err := ApplyTemplate("app_{tag}_{{ .ArchAlias }}.tar.gz", vars, "test-installer")
	assert.NoError(t, err)
	assert.Equal(t, "app_v1.0.0_arm64.tar.gz", result)
}

func TestApplyTemplateInvalidGoTemplate(t *testing.T) {
	logger.InitLogger(false)

	vars := NewTemplateVars("v1.0.0")

	// Invalid Go template syntax should return an error
	_, err := ApplyTemplate("app_{{ .InvalidField }.tar.gz", vars, "test-installer")
	assert.Error(t, err)
}

func TestArchDetection(t *testing.T) {
	logger.InitLogger(false)

	tests := []struct {
		goarch        string
		expectedArch  platform.Architecture
		expectedAlias string
		expectedGnu   string
	}{
		{"amd64", platform.ArchAmd64, "x86_64", "x86_64"},
		{"x86_64", platform.ArchAmd64, "x86_64", "x86_64"},
		{"arm64", platform.ArchArm64, "arm64", "aarch64"},
		{"aarch64", platform.ArchArm64, "arm64", "aarch64"},
	}

	for _, tc := range tests {
		t.Run(tc.goarch, func(t *testing.T) {
			platform.SetArch(tc.goarch)
			assert.Equal(t, tc.expectedArch, platform.GetArch())
			assert.Equal(t, tc.expectedAlias, platform.GetArchAlias())
			assert.Equal(t, tc.expectedGnu, platform.GetArchGnu())
		})
	}

	// Restore
	platform.SetArch("arm64")
}
