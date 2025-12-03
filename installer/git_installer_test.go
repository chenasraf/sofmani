package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
)

func newTestGitInstaller(data *appconfig.InstallerData) *GitInstaller {
	return &GitInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Info: data,
	}
}

func TestGitValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid: Both destination and ref are present
	validData := &appconfig.InstallerData{
		Name: strPtr("test-git-valid"),
		Type: appconfig.InstallerTypeGit,
		Opts: &map[string]any{
			"destination": "/some/path",
			"ref":         "main",
		},
	}
	assertNoValidationErrors(t, newTestGitInstaller(validData).Validate())

	// 🟢 Valid: Missing ref
	missingRefData := &appconfig.InstallerData{
		Name: strPtr("test-git-missing-ref"),
		Type: appconfig.InstallerTypeGit,
		Opts: &map[string]any{
			"destination": "/some/path",
		},
	}
	assertNoValidationErrors(t, newTestGitInstaller(missingRefData).Validate())

	// 🔴 Invalid: Missing destination
	missingDestData := &appconfig.InstallerData{
		Name: strPtr("test-git-missing-destination"),
		Type: appconfig.InstallerTypeGit,
		Opts: &map[string]any{
			"ref": "main",
		},
	}
	assertValidationError(t, newTestGitInstaller(missingDestData).Validate(), "destination")
}

func TestGitGetOpts(t *testing.T) {
	logger.InitLogger(false)

	// Test default opts (only destination set)
	defaultData := &appconfig.InstallerData{
		Name: strPtr("owner/repo"),
		Type: appconfig.InstallerTypeGit,
		Opts: &map[string]any{
			"destination": "/some/path",
		},
	}
	installer := newTestGitInstaller(defaultData)
	opts := installer.GetOpts()
	if opts.Destination == nil || *opts.Destination != "/some/path" {
		t.Errorf("expected Destination to be '/some/path'")
	}
	if opts.Ref != nil {
		t.Errorf("expected Ref to be nil")
	}
	if opts.Flags != nil {
		t.Errorf("expected Flags to be nil")
	}
	if opts.InstallFlags != nil {
		t.Errorf("expected InstallFlags to be nil")
	}
	if opts.UpdateFlags != nil {
		t.Errorf("expected UpdateFlags to be nil")
	}

	// Test with flags option
	flagsData := &appconfig.InstallerData{
		Name: strPtr("owner/repo"),
		Type: appconfig.InstallerTypeGit,
		Opts: &map[string]any{
			"destination": "/some/path",
			"flags":       "--depth 1",
		},
	}
	installerWithFlags := newTestGitInstaller(flagsData)
	optsWithFlags := installerWithFlags.GetOpts()
	if optsWithFlags.Flags == nil || *optsWithFlags.Flags != "--depth 1" {
		t.Errorf("expected Flags to be '--depth 1'")
	}

	// Test with install_flags option (for git clone)
	installFlagsData := &appconfig.InstallerData{
		Name: strPtr("owner/repo"),
		Type: appconfig.InstallerTypeGit,
		Opts: &map[string]any{
			"destination":   "/some/path",
			"install_flags": "--depth 1 --single-branch",
		},
	}
	installerWithInstallFlags := newTestGitInstaller(installFlagsData)
	optsWithInstallFlags := installerWithInstallFlags.GetOpts()
	if optsWithInstallFlags.InstallFlags == nil || *optsWithInstallFlags.InstallFlags != "--depth 1 --single-branch" {
		t.Errorf("expected InstallFlags to be '--depth 1 --single-branch'")
	}

	// Test with update_flags option (for git pull)
	updateFlagsData := &appconfig.InstallerData{
		Name: strPtr("owner/repo"),
		Type: appconfig.InstallerTypeGit,
		Opts: &map[string]any{
			"destination":  "/some/path",
			"update_flags": "--rebase",
		},
	}
	installerWithUpdateFlags := newTestGitInstaller(updateFlagsData)
	optsWithUpdateFlags := installerWithUpdateFlags.GetOpts()
	if optsWithUpdateFlags.UpdateFlags == nil || *optsWithUpdateFlags.UpdateFlags != "--rebase" {
		t.Errorf("expected UpdateFlags to be '--rebase'")
	}

	// Test with all options combined
	allOptsData := &appconfig.InstallerData{
		Name: strPtr("owner/repo"),
		Type: appconfig.InstallerTypeGit,
		Opts: &map[string]any{
			"destination":   "/some/path",
			"ref":           "develop",
			"flags":         "--common",
			"install_flags": "--depth 1",
			"update_flags":  "--rebase",
		},
	}
	installerWithAllOpts := newTestGitInstaller(allOptsData)
	optsWithAllOpts := installerWithAllOpts.GetOpts()
	if optsWithAllOpts.Destination == nil || *optsWithAllOpts.Destination != "/some/path" {
		t.Errorf("expected Destination to be '/some/path'")
	}
	if optsWithAllOpts.Ref == nil || *optsWithAllOpts.Ref != "develop" {
		t.Errorf("expected Ref to be 'develop'")
	}
	if optsWithAllOpts.Flags == nil || *optsWithAllOpts.Flags != "--common" {
		t.Errorf("expected Flags to be '--common'")
	}
	if optsWithAllOpts.InstallFlags == nil || *optsWithAllOpts.InstallFlags != "--depth 1" {
		t.Errorf("expected InstallFlags to be '--depth 1'")
	}
	if optsWithAllOpts.UpdateFlags == nil || *optsWithAllOpts.UpdateFlags != "--rebase" {
		t.Errorf("expected UpdateFlags to be '--rebase'")
	}
}
