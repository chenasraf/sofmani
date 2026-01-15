package installer

import (
	"fmt"
	"io"
	"maps"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/summary"
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
	// childResults stores results from nested installers.
	childResults []summary.InstallResult
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
	i.childResults = []summary.InstallResult{}
	for _, step := range config.Install {
		logger.Debug("Checking step %s", *step.Name)
		installer, err := GetInstaller(config, &step)
		if err != nil {
			return err
		}
		if installer == nil {
			logger.Warn("Installer type %s is not supported, skipping", step.Type)
		} else {
			result, err := RunInstaller(config, installer)
			if err != nil {
				logger.Error("Failed to run installer for step %s: %v", *step.Name, err)
				return fmt.Errorf("failed to run installer for step %s: %w", *step.Name, err)
			}
			if result != nil {
				i.childResults = append(i.childResults, *result)
			}
		}
	}
	return nil
}

// GetChildResults implements IChildResultsProvider.
func (i *ManifestInstaller) GetChildResults() []summary.InstallResult {
	return i.childResults
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
// It handles local files, Git repository URLs, and raw HTTP URLs.
func (i *ManifestInstaller) FetchManifest() error {
	opts := i.GetOpts()
	source := *opts.Source
	env := i.GetData().Environ()

	var config *appconfig.AppConfig
	var err error

	switch {
	case utils.IsGitURL(source):
		// Git repository URL - convert to raw URL and fetch
		content, fetchErr := i.getGitManifestConfig(source)
		if fetchErr != nil {
			return fetchErr
		}
		config, err = appconfig.ParseConfigFromContent([]byte(content))
		if err != nil {
			return fmt.Errorf("failed to parse manifest content: %w", err)
		}
	case strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://"):
		// Direct HTTP URL - fetch directly
		content, fetchErr := i.fetchRawURL(source)
		if fetchErr != nil {
			return fetchErr
		}
		config, err = appconfig.ParseConfigFromContent([]byte(content))
		if err != nil {
			return fmt.Errorf("failed to parse manifest content: %w", err)
		}
	default:
		// Local file path
		source = utils.GetRealPath(env, source)
		var path string
		if opts.Path == nil {
			path = ""
		} else {
			path = *opts.Path
		}
		path = utils.GetRealPath(env, path)
		fullPath := filepath.Join(source, path)
		logger.Debug("Parsing manifest from %s", fullPath)
		config, err = i.getLocalManifestConfig(fullPath)
		if err != nil {
			return err
		}
	}

	logger.Debug("Installers: %d", len(config.Install))
	config = i.inheritManifest(config)
	i.ManifestConfig = config
	return nil
}

// fetchRawURL fetches content directly from a raw HTTP URL.
func (i *ManifestInstaller) fetchRawURL(url string) (string, error) {
	logger.Debug("Fetching manifest from raw URL: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logger.Warn("failed to close response body: %v", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch manifest: HTTP %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read manifest content: %w", err)
	}

	return string(content), nil
}

func (i *ManifestInstaller) getGitManifestConfig(source string) (string, error) {
	opts := i.GetOpts()

	ref := "main"
	if opts.Ref != nil && *opts.Ref != "" {
		ref = *opts.Ref
	}

	path := ""
	if opts.Path != nil {
		path = *opts.Path
	}

	rawURL, err := utils.GetRawFileURL(source, ref, path)
	if err != nil {
		return "", fmt.Errorf("failed to construct raw file URL: %w", err)
	}

	logger.Debug("Fetching manifest from %s", rawURL)
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logger.Warn("failed to close response body: %v", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch manifest: HTTP %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read manifest content: %w", err)
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
