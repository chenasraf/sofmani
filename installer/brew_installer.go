package installer

import (
	"encoding/json"
	"slices"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type BrewInstaller struct {
	Config *appconfig.AppConfig
	Info   *appconfig.Installer
}

type BrewOpts struct {
	Tap         *string
	BinName     *string
	PreCommand  *string
	PostCommand *string
}

// Install implements IInstaller.
func (i *BrewInstaller) Install() error {
	chain := [][]string{
		{"brew", "install", i.Info.Name},
	}
	if i.GetOpts().PreCommand != nil {
		chain = append([][]string{{"sh", "-c", *i.GetOpts().PreCommand}}, chain...)
	}
	if i.GetOpts().PostCommand != nil {
		chain = append(chain, []string{"sh", "-c", *i.GetOpts().PostCommand})
	}
	return utils.RunCmdPassThroughChained(chain)
}

// Update implements IInstaller.
func (i *BrewInstaller) Update() error {
	return utils.RunCmdPassThrough("brew", "upgrade", i.Info.Name)
}

// CheckNeedsUpdate implements IInstaller.
func (i *BrewInstaller) CheckNeedsUpdate() (error, bool) {
	out, err := utils.RunCmdGetOutput("brew", "outdated", "--json", i.Info.Name)
	if err != nil {
		return err, false
	}
	jsonOut := make(map[string]interface{})
	err = json.Unmarshal(out, &jsonOut)
	if err != nil {
		return err, false
	}
	var formulae []interface{} = jsonOut["formulae"].([]interface{})
	strFormulae := make([]string, len(formulae))
	for i, v := range formulae {
		strFormulae[i] = v.(string)
	}
	if slices.Contains(strFormulae, i.Info.Name) {
		return nil, true
	}
	return nil, false
}

// CheckIsInstalled implements IInstaller.
func (i *BrewInstaller) CheckIsInstalled() (error, bool) {
	return utils.RunCmdGetSuccess("which", i.GetBinName())
}

// GetInfo implements IInstaller.
func (i *BrewInstaller) GetInfo() *appconfig.Installer {
	return i.Info
}

func (i *BrewInstaller) GetOpts() *BrewOpts {
	opts := &BrewOpts{}
	info := i.Info
	if info.Opts != nil {
		if tap, ok := (*info.Opts)["tap"].(string); ok {
			opts.Tap = &tap
		}
		if binName, ok := (*info.Opts)["bin_name"].(string); ok {
			opts.BinName = &binName
		}
		if command, ok := (*info.Opts)["pre_command"].(string); ok {
			opts.PreCommand = &command
		}
		if command, ok := (*info.Opts)["post_command"].(string); ok {
			opts.PostCommand = &command
		}
	}
	return opts
}

func (i *BrewInstaller) GetBinName() string {
	opts := i.GetOpts()
	if opts.BinName != nil && len(*opts.BinName) > 0 {
		return *opts.BinName
	}
	return i.Info.Name
}

func NewBrewInstaller(cfg *appconfig.AppConfig, installer *appconfig.Installer) *BrewInstaller {
	return &BrewInstaller{
		Config: cfg,
		Info:   installer,
	}
}
