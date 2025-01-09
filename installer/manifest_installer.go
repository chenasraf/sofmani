package installer

import (
	"fmt"
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
	info := i.GetInfo()
	if info.CheckHasUpdate != nil {
		return utils.RunCmdGetSuccess(info.Environ(), utils.GetOSShell(info.EnvShell), utils.GetOSShellArgs(*info.CheckHasUpdate)...)
	}
	return nil, true
}

// CheckIsInstalled implements IInstaller.
func (i *ManifestInstaller) CheckIsInstalled() (error, bool) {
	info := i.GetInfo()
	if info.CheckInstalled != nil {
		return utils.RunCmdGetSuccess(info.Environ(), utils.GetOSShell(info.EnvShell), utils.GetOSShellArgs(*info.CheckInstalled)...)
	}
	return nil, false
}

// GetInfo implements IInstaller.
func (i *ManifestInstaller) GetInfo() *appconfig.Installer {
	return i.Info
}

func (i *ManifestInstaller) GetOpts() *ManifestOpts {
	opts := &ManifestOpts{}
	info := i.GetInfo()
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

func (i *ManifestInstaller) FetchManifest() error {
	opts := i.GetOpts()
	source := *opts.Source
	isGit := utils.IsGitURL(source)
	env := i.GetInfo().Environ()
	var path string
	if opts.Path == nil {
		path = ""
	} else {
		path = *opts.Path
	}
	path = utils.GetRealPath(env, path)

	if isGit {
		src, err := i.getGitManifestConfig(source)
		if err != nil {
			return err
		}
		source = src
	} else {
		source = utils.GetRealPath(env, source)
	}

	logger.Debug("Parsing manifest from %s", filepath.Join(source, path))
	config, err := i.getLocalManifestConfig(filepath.Join(source, path))
	if err != nil {
		return err
	}
	logger.Debug("Installers: %d", len(config.Install))
	i.ManifestConfig = config
	return nil
}

func (i *ManifestInstaller) getGitManifestConfig(source string) (string, error) {
	opts := i.GetOpts()
	info := i.GetInfo()
	tmpDir, err := os.MkdirTemp("", "sofmani")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		return "", err
	}
	logger.Debug("Cloning %s to %s", source, tmpDir)
	err, success := utils.RunCmdGetSuccess(info.Environ(), "git", "clone", "--depth=1", source, tmpDir)
	if opts.Ref != nil {
		logger.Debug("Checking out ref %s", *opts.Ref)
		err = utils.RunCmdPassThrough(info.Environ(), "git", "-C", tmpDir, "checkout", *opts.Ref)
		if err != nil {
			return "", err
		}
	}
	if err != nil {
		return "", err
	}
	if success {
		return tmpDir, nil
	}
	return "", fmt.Errorf("Failed to clone %s", source)
}

func (i *ManifestInstaller) getLocalManifestConfig(path string) (*appconfig.AppConfig, error) {
	config, err := appconfig.ParseConfigFrom(path)

	if err != nil {
		return nil, err
	}

	logger.Debug("Setting manifest config")
	config = i.inheritManifest(config)
	return config, nil
}

func (i *ManifestInstaller) inheritManifest(config *appconfig.AppConfig) *appconfig.AppConfig {
	self := i.Config
	if self.Debug {
		config.Debug = self.Debug
	}
	if self.CheckUpdates {
		config.CheckUpdates = self.CheckUpdates
	}
	if self.Env != nil {
		logger.Debug("Injecting base env variables")
		var env map[string]string
		if config.Env == nil {
			env = make(map[string]string)
		} else {
			env = *config.Env
		}
		for k, v := range *self.Env {
			env[k] = v
		}
	}
	if self.Defaults != nil {
		defs := self.Defaults
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
	return config
}

func NewManifestInstaller(cfg *appconfig.AppConfig, installer *appconfig.Installer) *ManifestInstaller {
	return &ManifestInstaller{
		Config: cfg,
		Info:   installer,
	}
}
