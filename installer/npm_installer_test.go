package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
)

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
	assertNoValidationErrors(t, newTestNpmInstaller(validData).Validate())

	// 🔴 Edge case: nil name (will panic or fail in BaseValidate if implemented to check it)
	nilNameData := &appconfig.InstallerData{
		Name: nil,
		Type: appconfig.InstallerTypeNpm,
	}
	assertValidationError(t, newTestNpmInstaller(nilNameData).Validate(), "name")
}
