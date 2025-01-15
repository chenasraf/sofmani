package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
	"github.com/stretchr/testify/assert"
)

type MockInstaller struct {
	info         *appconfig.InstallerData
	isInstalled  bool
	needsUpdate  bool
	installError error
	updateError  error
	checkInstall error
	checkUpdate  error
}

func (m *MockInstaller) GetData() *appconfig.InstallerData {
	return m.info
}

func (m *MockInstaller) CheckIsInstalled() (error, bool) {
	return m.checkInstall, m.isInstalled
}

func (m *MockInstaller) CheckNeedsUpdate() (error, bool) {
	return m.checkUpdate, m.needsUpdate
}

func (m *MockInstaller) Install() error {
	return m.installError
}

func (m *MockInstaller) Update() error {
	return m.updateError
}

func TestGetInstaller(t *testing.T) {
	config := &appconfig.AppConfig{}
	logger.InitLogger(config.Debug)
	installer := &appconfig.InstallerData{Type: appconfig.InstallerTypeBrew}
	err, inst := GetInstaller(config, installer)
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
		info:        &appconfig.InstallerData{Name: strPtr("test"), Type: appconfig.InstallerTypeBrew},
		isInstalled: false,
	}
	err := RunInstaller(config, mockInstaller)
	assert.NoError(t, err)
}

func TestGetShouldRunOnOS(t *testing.T) {
	installer := &MockInstaller{
		info: &appconfig.InstallerData{
			Platforms: &platform.Platforms{
				Only: &[]platform.Platform{platform.PlatformMacos},
			},
		},
	}
	assert.True(t, installer.GetData().Platforms.GetShouldRunOnOS(platform.PlatformMacos))
	assert.False(t, installer.GetData().Platforms.GetShouldRunOnOS(platform.PlatformLinux))
}

func strPtr(s string) *string {
	return &s
}
