package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/stretchr/testify/assert"
)

type MockInstaller struct {
	data         *appconfig.InstallerData
	isInstalled  bool
	needsUpdate  bool
	installError error
	updateError  error
	checkInstall error
	checkUpdate  error
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

func strPtr(s string) *string {
	return &s
}
