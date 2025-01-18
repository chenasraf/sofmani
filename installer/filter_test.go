package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
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
				data: &appconfig.InstallerData{Name: strPtr("test"), Tags: strPtr("tag1")},
			},
			filters:  []string{},
			expected: true,
		},
		{
			name: "Positive filter match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: strPtr("test"), Tags: strPtr("tag1")},
			},
			filters:  []string{"test"},
			expected: true,
		},
		{
			name: "Positive filter no match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: strPtr("test"), Tags: strPtr("tag1")},
			},
			filters:  []string{"example"},
			expected: false,
		},
		{
			name: "Negative filter match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: strPtr("test"), Tags: strPtr("tag1")},
			},
			filters:  []string{"!test"},
			expected: false,
		},
		{
			name: "Negative filter no match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: strPtr("test"), Tags: strPtr("tag1")},
			},
			filters:  []string{"!example"},
			expected: true,
		},
		{
			name: "Tag filter match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: strPtr("test"), Tags: strPtr("tag1")},
			},
			filters:  []string{"tag:tag1"},
			expected: true,
		},
		{
			name: "Tag filter no match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: strPtr("test"), Tags: strPtr("tag1")},
			},
			filters:  []string{"tag:tag2"},
			expected: false,
		},
		{
			name: "Type filter match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: strPtr("test"), Type: appconfig.InstallerTypeBrew},
			},
			filters:  []string{"type:brew"},
			expected: true,
		},
		{
			name: "Type filter no match",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Name: strPtr("test"), Type: appconfig.InstallerTypeBrew},
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
				data: &appconfig.InstallerData{Enabled: strPtr("true")},
			},
			expected: true,
		},
		{
			name: "Enabled is false",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Enabled: strPtr("false")},
			},
			expected: false,
		},
		{
			name: "Enabled is a command that succeeds",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Enabled: strPtr("exit 0")},
			},
			expected: true,
		},
		{
			name: "Enabled is a command that fails",
			installer: &MockInstaller{
				data: &appconfig.InstallerData{Enabled: strPtr("exit 1")},
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
