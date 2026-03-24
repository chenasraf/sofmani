package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestFilterInstaller(t *testing.T) {
	tests := []struct {
		name      string
		installer IInstaller
		filters   []string
		expected  bool
	}{
		{
			name: "No filters",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Tags: lo.ToPtr("tag1")},
			},
			filters:  []string{},
			expected: true,
		},
		{
			name: "Positive filter match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Tags: lo.ToPtr("tag1")},
			},
			filters:  []string{"test"},
			expected: true,
		},
		{
			name: "Positive filter no match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Tags: lo.ToPtr("tag1")},
			},
			filters:  []string{"example"},
			expected: false,
		},
		{
			name: "Negative filter match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Tags: lo.ToPtr("tag1")},
			},
			filters:  []string{"!test"},
			expected: false,
		},
		{
			name: "Negative filter no match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Tags: lo.ToPtr("tag1")},
			},
			filters:  []string{"!example"},
			expected: true,
		},
		{
			name: "Tag filter match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Tags: lo.ToPtr("tag1")},
			},
			filters:  []string{"tag:tag1"},
			expected: true,
		},
		{
			name: "Tag filter no match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Tags: lo.ToPtr("tag1")},
			},
			filters:  []string{"tag:tag2"},
			expected: false,
		},
		{
			name: "Type filter match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Type: appconfig.InstallerTypeBrew},
			},
			filters:  []string{"type:brew"},
			expected: true,
		},
		{
			name: "Type filter no match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Type: appconfig.InstallerTypeBrew},
			},
			filters:  []string{"type:npm"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterInstaller(tt.installer, tt.filters)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInstallerIsEnabled(t *testing.T) {
	logger.InitLogger(true)
	tests := []struct {
		name      string
		installer IInstaller
		expected  bool
		expectErr bool
	}{
		{
			name: "Enabled is nil",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Enabled: nil},
			},
			expected: true,
		},
		{
			name: "Enabled is true",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Enabled: lo.ToPtr("true")},
			},
			expected: true,
		},
		{
			name: "Enabled is false",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Enabled: lo.ToPtr("false")},
			},
			expected: false,
		},
		{
			name: "Enabled is a command that succeeds",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Enabled: lo.ToPtr("exit 0")},
			},
			expected: true,
		},
		{
			name: "Enabled is a command that fails",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Enabled: lo.ToPtr("exit 1")},
			},
			expected:  false,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InstallerIsEnabled(tt.installer)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterInstallerEdgeCases(t *testing.T) {
	logger.InitLogger(false)

	t.Run("Multiple tags - one matches", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Tags: lo.ToPtr("tag1 tag2 tag3")},
		}
		result := FilterInstaller(installer, []string{"tag:tag2"})
		assert.True(t, result)
	})

	t.Run("Multiple tags - none match", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Tags: lo.ToPtr("tag1 tag2 tag3")},
		}
		result := FilterInstaller(installer, []string{"tag:tag4"})
		assert.False(t, result)
	})

	t.Run("Mixed positive and negative filters - positive matches, negative matches too", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Tags: lo.ToPtr("dev prod")},
		}
		// Should be excluded because negative filter matches
		result := FilterInstaller(installer, []string{"tag:dev", "!tag:prod"})
		assert.False(t, result)
	})

	t.Run("Mixed positive and negative filters - only positive matches", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Tags: lo.ToPtr("dev")},
		}
		result := FilterInstaller(installer, []string{"tag:dev", "!tag:prod"})
		assert.True(t, result)
	})

	t.Run("Name substring match", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("my-awesome-tool"), Tags: lo.ToPtr("")},
		}
		result := FilterInstaller(installer, []string{"awesome"})
		assert.True(t, result)
	})

	t.Run("Name exact match", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("vim"), Tags: lo.ToPtr("")},
		}
		result := FilterInstaller(installer, []string{"vim"})
		assert.True(t, result)
	})

	t.Run("Type filter case insensitive", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Type: appconfig.InstallerTypeBrew},
		}
		result := FilterInstaller(installer, []string{"type:BREW"})
		assert.True(t, result)
	})

	t.Run("Multiple positive filters - one matches", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("vim"), Tags: lo.ToPtr("editor")},
		}
		result := FilterInstaller(installer, []string{"neovim", "vim", "emacs"})
		assert.True(t, result)
	})

	t.Run("Multiple positive filters - none match", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("vim"), Tags: lo.ToPtr("editor")},
		}
		result := FilterInstaller(installer, []string{"neovim", "emacs", "nano"})
		assert.False(t, result)
	})

	t.Run("Multiple negative filters - all don't match", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("vim"), Tags: lo.ToPtr("editor")},
		}
		result := FilterInstaller(installer, []string{"!neovim", "!emacs"})
		assert.True(t, result)
	})

	t.Run("Multiple negative filters - one matches", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("vim"), Tags: lo.ToPtr("editor")},
		}
		result := FilterInstaller(installer, []string{"!neovim", "!vim"})
		assert.False(t, result)
	})

	t.Run("Negative type filter", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Type: appconfig.InstallerTypeBrew},
		}
		result := FilterInstaller(installer, []string{"!type:npm"})
		assert.True(t, result)
	})

	t.Run("Negative type filter matches", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Type: appconfig.InstallerTypeBrew},
		}
		result := FilterInstaller(installer, []string{"!type:brew"})
		assert.False(t, result)
	})

	t.Run("Negative tag filter", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Tags: lo.ToPtr("dev")},
		}
		result := FilterInstaller(installer, []string{"!tag:prod"})
		assert.True(t, result)
	})

	t.Run("Negative tag filter matches", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Name: lo.ToPtr("test"), Tags: lo.ToPtr("dev")},
		}
		result := FilterInstaller(installer, []string{"!tag:dev"})
		assert.False(t, result)
	})
}

func TestInstallerIsEnabledEdgeCases(t *testing.T) {
	logger.InitLogger(false)

	t.Run("Enabled is TRUE (uppercase)", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Enabled: lo.ToPtr("TRUE")},
		}
		result, err := InstallerIsEnabled(installer)
		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("Enabled is FALSE (uppercase)", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Enabled: lo.ToPtr("FALSE")},
		}
		result, err := InstallerIsEnabled(installer)
		assert.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("Enabled is True (mixed case)", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Enabled: lo.ToPtr("True")},
		}
		result, err := InstallerIsEnabled(installer)
		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("Enabled is a command that checks for which", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Enabled: lo.ToPtr("which which")},
		}
		result, err := InstallerIsEnabled(installer)
		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("Enabled is a command that checks for nonexistent binary", func(t *testing.T) {
		installer := &MockInstaller{
			data: &appconfig.InstallerData{Enabled: lo.ToPtr("which nonexistent-binary-12345")},
		}
		result, err := InstallerIsEnabled(installer)
		assert.NoError(t, err)
		assert.False(t, result)
	})
}
