package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/samber/lo"
)

func newTestGoInstaller(data *appconfig.InstallerData) *GoInstaller {
	return &GoInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config: nil,
		Info:   data,
	}
}

func TestGoValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid go installer
	validData := &appconfig.InstallerData{
		Name: lo.ToPtr("golang.org/x/tools/gopls"),
		Type: appconfig.InstallerTypeGo,
	}
	assertNoValidationErrors(t, newTestGoInstaller(validData).Validate())

	// 🔴 Nil name
	nilNameData := &appconfig.InstallerData{
		Name: nil,
		Type: appconfig.InstallerTypeGo,
	}
	assertValidationError(t, newTestGoInstaller(nilNameData).Validate(), "name")
}

func TestGoGetOpts(t *testing.T) {
	logger.InitLogger(false)

	// Default opts
	defaultData := &appconfig.InstallerData{
		Name: lo.ToPtr("golang.org/x/tools/gopls"),
		Type: appconfig.InstallerTypeGo,
	}
	installer := newTestGoInstaller(defaultData)
	opts := installer.GetOpts()
	if opts.Version != nil {
		t.Errorf("expected Version to be nil")
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

	allOptsData := &appconfig.InstallerData{
		Name: lo.ToPtr("golang.org/x/tools/gopls"),
		Type: appconfig.InstallerTypeGo,
		Opts: &map[string]any{
			"version":       "v0.16.0",
			"flags":         "--common",
			"install_flags": "--install-specific",
			"update_flags":  "--update-specific",
		},
	}
	installerAll := newTestGoInstaller(allOptsData)
	optsAll := installerAll.GetOpts()
	if optsAll.Version == nil || *optsAll.Version != "v0.16.0" {
		t.Errorf("expected Version to be 'v0.16.0'")
	}
	if optsAll.Flags == nil || *optsAll.Flags != "--common" {
		t.Errorf("expected Flags to be '--common'")
	}
	if optsAll.InstallFlags == nil || *optsAll.InstallFlags != "--install-specific" {
		t.Errorf("expected InstallFlags to be '--install-specific'")
	}
	if optsAll.UpdateFlags == nil || *optsAll.UpdateFlags != "--update-specific" {
		t.Errorf("expected UpdateFlags to be '--update-specific'")
	}
}

func TestGoGetPackageRef(t *testing.T) {
	logger.InitLogger(false)

	// Default: appends @latest
	data := &appconfig.InstallerData{
		Name: lo.ToPtr("golang.org/x/tools/gopls"),
		Type: appconfig.InstallerTypeGo,
	}
	if ref := newTestGoInstaller(data).GetPackageRef(); ref != "golang.org/x/tools/gopls@latest" {
		t.Errorf("expected default ref to be 'golang.org/x/tools/gopls@latest', got %q", ref)
	}

	// opts.version overrides default
	versionData := &appconfig.InstallerData{
		Name: lo.ToPtr("golang.org/x/tools/gopls"),
		Type: appconfig.InstallerTypeGo,
		Opts: &map[string]any{"version": "v0.16.0"},
	}
	if ref := newTestGoInstaller(versionData).GetPackageRef(); ref != "golang.org/x/tools/gopls@v0.16.0" {
		t.Errorf("expected ref with opts.version, got %q", ref)
	}

	// Inline @version on name wins
	inlineData := &appconfig.InstallerData{
		Name: lo.ToPtr("golang.org/x/tools/gopls@v0.15.0"),
		Type: appconfig.InstallerTypeGo,
		Opts: &map[string]any{"version": "v0.16.0"},
	}
	if ref := newTestGoInstaller(inlineData).GetPackageRef(); ref != "golang.org/x/tools/gopls@v0.15.0" {
		t.Errorf("expected inline @version to take precedence, got %q", ref)
	}
}

func TestGoGetBinName(t *testing.T) {
	logger.InitLogger(false)

	// Default: last path component of name
	defaultData := &appconfig.InstallerData{
		Name: lo.ToPtr("golang.org/x/tools/gopls"),
		Type: appconfig.InstallerTypeGo,
	}
	if name := newTestGoInstaller(defaultData).GetBinName(); name != "gopls" {
		t.Errorf("expected bin name 'gopls', got %q", name)
	}

	// Inline @version is stripped
	withVersion := &appconfig.InstallerData{
		Name: lo.ToPtr("golang.org/x/tools/gopls@v0.16.0"),
		Type: appconfig.InstallerTypeGo,
	}
	if name := newTestGoInstaller(withVersion).GetBinName(); name != "gopls" {
		t.Errorf("expected bin name 'gopls', got %q", name)
	}

	// Override via bin_name
	override := &appconfig.InstallerData{
		Name:    lo.ToPtr("golang.org/x/tools/gopls"),
		Type:    appconfig.InstallerTypeGo,
		BinName: lo.ToPtr("gopls-custom"),
	}
	if name := newTestGoInstaller(override).GetBinName(); name != "gopls-custom" {
		t.Errorf("expected bin name 'gopls-custom', got %q", name)
	}
}

func TestGoGetData(t *testing.T) {
	logger.InitLogger(false)

	data := &appconfig.InstallerData{
		Name: lo.ToPtr("golang.org/x/tools/gopls"),
		Type: appconfig.InstallerTypeGo,
	}
	installer := newTestGoInstaller(data)
	if installer.GetData() != data {
		t.Errorf("expected GetData to return the same data pointer")
	}
}
