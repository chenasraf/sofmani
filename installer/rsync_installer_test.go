package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/samber/lo"
)

func newTestRsyncInstaller(data *appconfig.InstallerData) *RsyncInstaller {
	return &RsyncInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config: nil,
		Info:   data,
	}
}

func TestRsyncValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid rsync config
	validData := &appconfig.InstallerData{
		Name: lo.ToPtr("rsync-valid"),
		Type: appconfig.InstallerTypeRsync,
		Opts: &map[string]any{
			"source":      "/path/from",
			"destination": "/path/to",
			"flags":       "-avz",
		},
	}
	assertNoValidationErrors(t, newTestRsyncInstaller(validData).Validate())

	// 🔴 Missing source
	missingSource := &appconfig.InstallerData{
		Name: lo.ToPtr("rsync-missing-source"),
		Type: appconfig.InstallerTypeRsync,
		Opts: &map[string]any{
			"destination": "/path/to",
		},
	}
	assertValidationError(t, newTestRsyncInstaller(missingSource).Validate(), "source")

	// 🔴 Missing destination
	missingDest := &appconfig.InstallerData{
		Name: lo.ToPtr("rsync-missing-destination"),
		Type: appconfig.InstallerTypeRsync,
		Opts: &map[string]any{
			"source": "/path/from",
		},
	}

	assertValidationError(t, newTestRsyncInstaller(missingDest).Validate(), "destination")

	// 🔴 Empty flags string
	emptyFlags := &appconfig.InstallerData{
		Name: lo.ToPtr("rsync-empty-flags"),
		Type: appconfig.InstallerTypeRsync,
		Opts: &map[string]any{
			"source":      "/path/from",
			"destination": "/path/to",
			"flags":       "",
		},
	}

	assertValidationError(t, newTestRsyncInstaller(emptyFlags).Validate(), "flags")
}
