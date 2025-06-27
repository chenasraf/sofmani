package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
)

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
	assertNoValidationErrors(t, newTestGroupInstaller(validData).Validate())

	// 🔴 Invalid: empty steps slice
	emptySteps := &appconfig.InstallerData{
		Name:  strPtr("group-empty"),
		Type:  appconfig.InstallerTypeGroup,
		Steps: &[]appconfig.InstallerData{},
	}
	assertValidationError(t, newTestGroupInstaller(emptySteps).Validate(), "steps")

	// 🔴 Invalid: nil steps
	nilSteps := &appconfig.InstallerData{
		Name:  strPtr("group-nil"),
		Type:  appconfig.InstallerTypeGroup,
		Steps: nil,
	}
	assertValidationError(t, newTestGroupInstaller(nilSteps).Validate(), "steps")
}
