package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
)

func newAptInstaller(data *appconfig.InstallerData) *AptInstaller {
	return &AptInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config: nil,
		Info:   data,
	}
}

func TestAptValidation(t *testing.T) {
	logger.InitLogger(false)
	aptInstaller := newAptInstaller(
		&appconfig.InstallerData{
			Name: strPtr("test-apt"),
			Type: appconfig.InstallerTypeApt,
		},
	)
	assertNoValidationErrors(t, aptInstaller.Validate())
}
