package installer

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/utils"
)

// ManifestInstaller is an installer that installs software based on another sofmani manifest file.
type ManifestInstaller struct {
	InstallerBase
	// Config is the main application configuration.
	Config *appconfig.AppConfig
	// Info is the installer data for this manifest installer.
	Info *appconfig.InstallerData
	// ManifestConfig is the configuration loaded from the manifest file.
	ManifestConfig *appconfig.AppConfig
}

// ManifestOpts represents options for the ManifestInstaller.
type ManifestOpts struct {
	// Source is the source of the manifest file. It can be a local path or a Git URL.
	Source *string
	// Path is the path to the manifest file within the source (if applicable, e.g., in a Git repository).
	Path *string
	// Ref is the Git reference (branch, tag, or commit) to use if the source is a Git URL.
	Ref *string
}

// Validate validates the installer configuration.
func (i *ManifestInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	info := i.GetData()
	opts := i.GetOpts()
	if opts.Source == nil || len(*opts.Source) == 0 {
		errors = append(errors, ValidationError{FieldName: "source", Message: validationIsRequired(), InstallerName: *info.Name})
	}
	if opts.Path == nil || len(*opts.Path) == 0 {
		errors = append(errors, ValidationError{FieldName: "path", Message: validationIsRequired(), InstallerName: *info.Name})
	}
	if opts.Ref != nil && len(*opts.Ref) == 0 {
		errors = append(errors, ValidationError{FieldName: "ref", Message: validationIsNotEmpty(), InstallerName: *info.Name})
	}
	return errors
}

// Install implements IInstaller.
func (i *ManifestInstaller) Install() error {
	logger.Debug("Getting manifest info...")
	err := i.FetchManifest()
	if err != nil {
		return err
	}
	info := i.GetData()
	name := *info.Name
	config := i.ManifestConfig
	logger.Info("Installing manifest %s", name)
	for _, step := range config.Install {
		logger.Debug("Checking step %s", *step.Name)
		installer, err := GetInstaller(config, &step)
		if err != nil {
			return err
		}
		if installer == nil {
			logger.Warn("Installer type %s is not supported, skipping", step.Type)
		} else {
			err := RunInstaller(config, installer)
			if err != nil {
				logger.Error("Failed to run installer for step %s: %v", *step.Name, err)
				return fmt.Errorf("failed to run installer for step %s: %w", *step.Name, err)
			}
		}
	}
	return nil
}

// Update implements IInstaller.
func (i *ManifestInstaller) Update() error {
	return i.Install()
}

// CheckNeedsUpdate implements IInstaller.
func (i *ManifestInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	return true, nil
}

// CheckIsInstalled implements IInstaller.
func (i *ManifestInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return false, nil
}

// GetData implements IInstaller.
func (i *ManifestInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

// GetOpts returns the parsed options for the ManifestInstaller.
func (i *ManifestInstaller) GetOpts() *ManifestOpts {
	opts := &ManifestOpts{}
	info := i.GetData()
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

// FetchManifest fetches and parses the manifest file.
// It handles both local and Git sources.
func (i *ManifestInstaller) FetchManifest() error {
	opts := i.GetOpts()
	source := *opts.Source
	isGit := utils.IsGitURL(source)
	env := i.GetData().Environ()
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
	tmpDir, err := os.MkdirTemp("", "sofmani")
	if err != nil {
		return "", err
	}

	defer func() {
		if rmErr := os.RemoveAll(tmpDir); rmErr != nil {
			logger.Warn("Failed to clean up tmp dir %s: %v", tmpDir, rmErr)
		}
	}()

	logger.Debug("Cloning %s to %s", source, tmpDir)
	success, err := i.RunCmdGetSuccess("git", "clone", "--depth=1", source, tmpDir)
	if err != nil {
		return "", err
	}

	if opts.Ref != nil {
		logger.Debug("Checking out ref %s", *opts.Ref)
		err = i.RunCmdPassThrough("git", "-C", tmpDir, "checkout", *opts.Ref)
		if err != nil {
			return "", err
		}
	}

	if !success {
		return "", fmt.Errorf("failed to clone %s", source)
	}

	contentPath := filepath.Join(tmpDir, "manifest.yaml")
	content, err := os.ReadFile(contentPath)
	if err != nil {
		return "", fmt.Errorf("failed to read manifest file: %w", err)
	}

	return string(content), nil
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
	if self.Debug != nil {
		config.Debug = self.Debug
	}
	if *self.CheckUpdates {
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
		maps.Copy(env, *self.Env)
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

func NewManifestInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *ManifestInstaller {
	return &ManifestInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Info:          installer,
	}
}
