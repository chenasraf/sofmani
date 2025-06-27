package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
)

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
	assertNoValidationErrors(t, newTestPipxInstaller(validData).Validate())

	// 🔴 Optional: test nil name if BaseValidate handles it
	nilNameData := &appconfig.InstallerData{
		Name: nil,
		Type: appconfig.InstallerTypePipx,
	}
	assertValidationError(t, newTestPipxInstaller(nilNameData).Validate(), "name")
}
