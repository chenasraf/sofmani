package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/samber/lo"
)

func newTestShellInstaller(data *appconfig.InstallerData) *ShellInstaller {
	return &ShellInstaller{
		InstallerBase: InstallerBase{Data: data},
		Config:        nil,
		Info:          data,
	}
}

func TestShellValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid shell config
	validData := &appconfig.InstallerData{
		Name: lo.ToPtr("shell-valid"),
		Type: appconfig.InstallerTypeShell,
		Opts: &map[string]any{
			"command":        "echo install",
			"update_command": "echo update",
		},
	}
	assertNoValidationErrors(t, newTestShellInstaller(validData).Validate())

	// 🔴 Missing command
	missingCommand := &appconfig.InstallerData{
		Name: lo.ToPtr("shell-missing-command"),
		Type: appconfig.InstallerTypeShell,
		Opts: &map[string]any{
			"update_command": "echo update",
		},
	}
	assertValidationError(t, newTestShellInstaller(missingCommand).Validate(), "command")

	// 🟢 Valid - missing update_command
	missingUpdate := &appconfig.InstallerData{
		Name: lo.ToPtr("shell-missing-update"),
		Type: appconfig.InstallerTypeShell,
		Opts: &map[string]any{
			"command": "echo install",
		},
	}
	assertNoValidationErrors(t, newTestShellInstaller(missingUpdate).Validate())

	// 🔴 Missing both
	missingBoth := &appconfig.InstallerData{
		Name: lo.ToPtr("shell-missing-both"),
		Type: appconfig.InstallerTypeShell,
		Opts: &map[string]any{},
	}
	assertValidationError(t, newTestShellInstaller(missingBoth).Validate(), "command")
}
