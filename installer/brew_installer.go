package installer

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"slices"

	"github.com/chenasraf/sofmani/appconfig"
)

type BrewInstaller struct {
	Config *appconfig.AppConfig
	Info   *appconfig.Installer
}

type BrewOpts struct {
	Tap     *string
	BinName *string
}

// Install implements IInstaller.
func (i *BrewInstaller) Install() error {
	cmd := exec.Command("brew", "install", i.Info.Name)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	cmd.Wait()
	return nil
}

// Update implements IInstaller.
func (i *BrewInstaller) Update() error {
	cmd := exec.Command("brew", "upgrade", i.Info.Name)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	cmd.Wait()
	return nil
}

// CheckNeedsUpdate implements IInstaller.
func (i *BrewInstaller) CheckNeedsUpdate() (error, bool) {
	cmd := exec.Command("brew", "outdated", "--json", i.Info.Name)
	out, err := cmd.Output()
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
	cmd := exec.Command("which", i.GetBinName())
	err := cmd.Run()
	if err != nil {
		return nil, false
	}
	return nil, true
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
