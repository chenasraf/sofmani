package installer

import (
	"os"
	"path/filepath"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/utils"
)

type ManifestInstaller struct {
	Config         *appconfig.AppConfig
	Info           *appconfig.Installer
	ManifestConfig *appconfig.AppConfig
}

type ManifestOpts struct {
	Source *string
	Path   *string
	Ref    *string
}

// Install implements IInstaller.
func (i *ManifestInstaller) Install() error {
	logger.Debug("Getting manifest info...")
	err := i.FetchManifest()
	if err != nil {
		return err
	}
	info := i.GetInfo()
	name := *info.Name
	config := i.ManifestConfig
	logger.Info("Installing manifest %s", name)
	for _, step := range config.Install {
		logger.Debug("Checking step %s", *step.Name)
		err, installer := GetInstaller(config, &step)
		if err != nil {
			return err
		}
		if installer == nil {
			logger.Warn("Installer type %s is not supported, skipping", step.Type)
		} else {
			RunInstaller(config, installer)
		}
	}
	return nil
}

// Update implements IInstaller.
func (i *ManifestInstaller) Update() error {
	return i.Install()
}

// CheckNeedsUpdate implements IInstaller.
func (i *ManifestInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetInfo().CheckHasUpdate != nil {
		return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetOSShell(i.GetInfo().EnvShell), utils.GetOSShellArgs(*i.GetInfo().CheckHasUpdate)...)
	}
	return nil, true
}

// CheckIsInstalled implements IInstaller.
func (i *ManifestInstaller) CheckIsInstalled() (error, bool) {
	if i.GetInfo().CheckInstalled != nil {
		return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetOSShell(i.GetInfo().EnvShell), utils.GetOSShellArgs(*i.GetInfo().CheckInstalled)...)
	}
	return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetShellWhich(), i.GetBinName())
}

// GetInfo implements IInstaller.
func (i *ManifestInstaller) GetInfo() *appconfig.Installer {
	return i.Info
}

func (i *ManifestInstaller) GetOpts() *ManifestOpts {
	opts := &ManifestOpts{}
	info := i.Info
	if info.Opts != nil {
		if source, ok := (*info.Opts)["source"].(string); ok {
			opts.Source = &source
		}
		if path, ok := (*info.Opts)["path"].(string); ok {
			opts.Path = &path
		}
		if ref, ok := (*info.Opts)["ref"].(string); ok {
			opts.Ref = &ref
		}
	}
	return opts
}

func (i *ManifestInstaller) GetBinName() string {
	info := i.GetInfo()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

func (i *ManifestInstaller) FetchManifest() error {
	opts := i.GetOpts()
	source := *opts.Source
	isGit := utils.IsGitURL(source)
	var path string
	if opts.Path == nil {
		path = ""
	} else {
		path = *opts.Path
	}
	path = utils.GetRealPath(i.GetInfo().Environ(), path)

	if isGit {
		tmpDir, err := os.MkdirTemp("", "sofmani")
		defer os.RemoveAll(tmpDir)
		if err != nil {
			return err
		}
		logger.Debug("Cloning %s to %s", source, tmpDir)
		err, success := utils.RunCmdGetSuccess(i.Info.Environ(), "git", "clone", "--depth=1", source, tmpDir)
		if opts.Ref != nil {
			logger.Debug("Checking out ref %s", *opts.Ref)
			err = utils.RunCmdPassThrough(i.Info.Environ(), "git", "-C", tmpDir, "checkout", *opts.Ref)
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
		if success {
			source = tmpDir
		}
	} else {
		source = utils.GetRealPath(i.GetInfo().Environ(), source)
	}

	logger.Debug("Parsing manifest from %s", filepath.Join(source, path))
	config, err := appconfig.ParseConfigFrom(filepath.Join(source, path))

	if err != nil {
		return err
	}

	logger.Debug("Setting manifest config")
	if i.Config.Debug {
		config.Debug = i.Config.Debug
	}
	if i.Config.CheckUpdates {
		config.CheckUpdates = i.Config.CheckUpdates
	}
	if i.Config.Env != nil {
		logger.Debug("Injecting base env variables")
		var env map[string]string
		if config.Env == nil {
			env = make(map[string]string)
		} else {
			env = *config.Env
		}
		for k, v := range *i.Config.Env {
			env[k] = v
		}
	}
	if i.Config.Defaults != nil {
		defs := i.Config.Defaults
		if defs.Type != nil {
			types := *defs.Type
			if shell, ok := types["shell"]; ok {
				logger.Debug("Setting shell to %v", shell)
				if config.Defaults == nil {
					config.Defaults = &appconfig.AppConfigDefaults{}
				}
				confDefs := *config.Defaults.Type
				confDefs["shell"] = shell
			}
		}
	}
	logger.Debug("Installers: %d", len(config.Install))
	i.ManifestConfig = config
	return nil
}

func NewManifestInstaller(cfg *appconfig.AppConfig, installer *appconfig.Installer) *ManifestInstaller {
	return &ManifestInstaller{
		Config: cfg,
		Info:   installer,
	}
}
