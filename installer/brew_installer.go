package installer

import (
	"fmt"
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type BrewInstaller struct {
	InstallerBase
	Config *appconfig.AppConfig
	Info   *appconfig.InstallerData
}

type BrewOpts struct {
	Tap *string
}

func (i *BrewInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	info := i.GetData()
	opts := i.GetOpts()
	if opts.Tap != nil {
		if !strings.Contains(*opts.Tap, "/") || len(*opts.Tap) < 3 {
			errors = append(errors, ValidationError{FieldName: "tap", Message: validationInvalidFormat(), InstallerName: *info.Name})
		}
	}
	return errors
}

// Install implements IInstaller.
func (i *BrewInstaller) Install() error {
	name := i.GetFullName()
	return i.RunCmdAsFile(fmt.Sprintf("brew install %s", name))
}

// Update implements IInstaller.
func (i *BrewInstaller) Update() error {
	name := i.GetFullName()
	return i.RunCmdAsFile(fmt.Sprintf("brew upgrade %s", name))
}

func (i *BrewInstaller) GetFullName() string {
	name := *i.Info.Name
	if i.GetOpts().Tap != nil {
		name = *i.GetOpts().Tap + "/" + name
	}
	return name
}

// CheckNeedsUpdate implements IInstaller.
func (i *BrewInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	name := i.GetFullName()
	cmd := fmt.Sprintf(
		`brew outdated --json %s %s`,
		name,
		PipedInputNeedsUpdateCommand,
	)
	success, err := i.RunCmdGetSuccessPassThrough("bash", "-c", cmd)
	if err != nil {
		return false, err
	}
	return !success, nil
}

const PipedInputNeedsUpdateCommand = `| awk '
  BEGIN { in_json = 0; json = "" }
  /^ *\{/ { in_json = 1 }
  in_json {
    json = json $0 ORS;
    if ($0 ~ /^\}/) {
      in_json = 0;
      next;
    }
    next;
  }
  { print }
  END {
    cleaned = json;
    gsub(/[[:space:]]/, "", cleaned);
    if (cleaned != "{\"formulae\":[],\"casks\":[]}") printf "%s", json;
  }
'`

// CheckIsInstalled implements IInstaller.
func (i *BrewInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return i.RunCmdGetSuccess(utils.GetShellWhich(), i.GetBinName())
}

// GetData implements IInstaller.
func (i *BrewInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

func (i *BrewInstaller) GetOpts() *BrewOpts {
	opts := &BrewOpts{}
	info := i.Info
	if info.Opts != nil {
		if tap, ok := (*info.Opts)["tap"].(string); ok {
			opts.Tap = &tap
		}
	}
	return opts
}

func (i *BrewInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

func NewBrewInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *BrewInstaller {
	i := &BrewInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Info:          installer,
	}

	return i
}
