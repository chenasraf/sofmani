package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
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
