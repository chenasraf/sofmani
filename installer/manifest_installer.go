package installer

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
	"github.com/chenasraf/sofmani/summary"
	"github.com/chenasraf/sofmani/utils"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
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
	// Inherit selects which global settings the loaded manifest receives from the config that loads it.
	Inherit ManifestInherit
	// Overrides holds global settings applied on top of everything else the loaded manifest ends up with.
	Overrides *ManifestSettings
}

// ManifestInherit selects which global settings a loaded manifest inherits. In YAML it is either a
// boolean covering every setting, or a map of setting name to boolean where omitted settings are
// not inherited. A manifest inherits nothing unless asked to, so a remote manifest never sees the
// loading config's env by accident.
type ManifestInherit struct {
	// Env covers both `env` and `platform_env`.
	Env            bool `yaml:"env"`
	Defaults       bool `yaml:"defaults"`
	RepoUpdate     bool `yaml:"repo_update"`
	MachineAliases bool `yaml:"machine_aliases"`
	CheckUpdates   bool `yaml:"check_updates"`
}

// inheritAll is what `inherit: true` selects.
var inheritAll = ManifestInherit{Env: true, Defaults: true, RepoUpdate: true, MachineAliases: true, CheckUpdates: true}

// ManifestSettings are the global settings a manifest installer can pass to the manifest it loads.
type ManifestSettings struct {
	CheckUpdates   *bool                                                 `yaml:"check_updates"`
	RepoUpdate     *map[appconfig.InstallerType]appconfig.RepoUpdateMode `yaml:"repo_update"`
	Defaults       *appconfig.AppConfigDefaults                          `yaml:"defaults"`
	Env            *map[string]string                                    `yaml:"env"`
	PlatformEnv    *platform.PlatformMap[map[string]string]              `yaml:"platform_env"`
	MachineAliases *map[string]string                                    `yaml:"machine_aliases"`
}

// decodeStrict re-encodes a loosely typed opts value and decodes it into out, rejecting unknown keys
// so that a misspelled setting fails validation instead of being silently inherited.
func decodeStrict(value any, out any) error {
	raw, err := yaml.Marshal(value)
	if err != nil {
		return err
	}
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	return dec.Decode(out)
}

// parseInherit reads `opts.inherit`.
func (i *ManifestInstaller) parseInherit() (ManifestInherit, error) {
	inherit := ManifestInherit{}
	if i.GetData().Opts == nil {
		return inherit, nil
	}
	value, ok := (*i.GetData().Opts)["inherit"]
	if !ok || value == nil {
		return inherit, nil
	}
	if all, ok := value.(bool); ok {
		if all {
			return inheritAll, nil
		}
		return ManifestInherit{}, nil
	}
	if _, ok := value.(map[string]any); !ok {
		return inherit, fmt.Errorf("must be a boolean or a map of setting names to booleans")
	}
	if err := decodeStrict(value, &inherit); err != nil {
		return inherit, err
	}
	return inherit, nil
}

// parseOverrides reads `opts.overrides`.
func (i *ManifestInstaller) parseOverrides() (*ManifestSettings, error) {
	if i.GetData().Opts == nil {
		return nil, nil
	}
	value, ok := (*i.GetData().Opts)["overrides"]
	if !ok || value == nil {
		return nil, nil
	}
	if _, ok := value.(map[string]any); !ok {
		return nil, fmt.Errorf("must be a map of global settings")
	}
	overrides := &ManifestSettings{}
	if err := decodeStrict(value, overrides); err != nil {
		return nil, err
	}
	return overrides, nil
}

// Validate validates the installer configuration.
func (i *ManifestInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	info := i.GetData()
	opts := i.GetOpts()
	if _, err := i.parseInherit(); err != nil {
		errors = append(errors, ValidationError{FieldName: "inherit", Message: err.Error(), InstallerName: *info.Name})
	}
	if _, err := i.parseOverrides(); err != nil {
		errors = append(errors, ValidationError{FieldName: "overrides", Message: err.Error(), InstallerName: *info.Name})
	}
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
	restoreEnv, err := i.scopeManifestEnv()
	if err != nil {
		return err
	}
	defer restoreEnv()
	logger.Info("Installing manifest %s", logger.H(name))
	i.childResults = []summary.InstallResult{}
	for _, step := range config.Install {
		logger.Debug("Checking step %s", logger.H(*step.Name))
		installer, err := GetInstaller(config, &step)
		if err != nil {
			return err
		}
		if installer == nil {
			logger.Warn("Installer type %s is not supported, skipping", logger.H(string(step.Type)))
		} else {
			result, err := RunInstaller(config, installer)
			if err != nil {
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
	// Invalid values are reported by Validate and refuse to install in FetchManifest.
	opts.Inherit, _ = i.parseInherit()
	opts.Overrides, _ = i.parseOverrides()
	return opts
}

// FetchManifest fetches and parses the manifest file.
// It handles local files, Git repository URLs, and raw HTTP URLs.
func (i *ManifestInstaller) FetchManifest() error {
	// An unreadable inherit or overrides must not fall back to inheriting everything.
	if _, err := i.parseInherit(); err != nil {
		return fmt.Errorf("invalid inherit for manifest %s: %w", *i.GetData().Name, err)
	}
	if _, err := i.parseOverrides(); err != nil {
		return fmt.Errorf("invalid overrides for manifest %s: %w", *i.GetData().Name, err)
	}
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
			return fmt.Errorf("failed to parse manifest content from %s: %w", source, err)
		}
	case strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://"):
		// Direct HTTP URL - fetch directly
		content, fetchErr := i.fetchRawURL(source)
		if fetchErr != nil {
			return fetchErr
		}
		config, err = appconfig.ParseConfigFromContent([]byte(content))
		if err != nil {
			return fmt.Errorf("failed to parse manifest content from %s: %w", source, err)
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
			return fmt.Errorf("failed to load manifest from %s: %w", fullPath, err)
		}
	}

	logger.Debug("Installers: %d", len(config.Install))
	i.ManifestConfig = i.buildManifestConfig(config, opts)
	return nil
}

// fetchRawURL fetches content directly from a raw HTTP URL.
func (i *ManifestInstaller) fetchRawURL(url string) (string, error) {
	logger.Debug("Fetching manifest from raw URL: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch manifest from %s: %w", url, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logger.Warn("failed to close response body: %v", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch manifest from %s: HTTP %d", url, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read manifest content from %s: %w", url, err)
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
		return "", fmt.Errorf("failed to construct raw file URL (source=%s, ref=%s, path=%s): %w", source, ref, path, err)
	}

	logger.Debug("Fetching manifest from %s", rawURL)
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch manifest from %s: %w", rawURL, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logger.Warn("failed to close response body: %v", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch manifest from %s: HTTP %d", rawURL, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read manifest content from %s: %w", rawURL, err)
	}

	return string(content), nil
}

func (i *ManifestInstaller) getLocalManifestConfig(path string) (*appconfig.AppConfig, error) {
	config, err := appconfig.ParseConfigFrom(path)

	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest at %s: %w", path, err)
	}

	return config, nil
}

// buildManifestConfig layers the global settings of a loaded manifest. From lowest to highest
// precedence: the manifest's own settings, the inherited settings of the loading config, and the
// installer's `overrides`.
func (i *ManifestInstaller) buildManifestConfig(config *appconfig.AppConfig, opts *ManifestOpts) *appconfig.AppConfig {
	if parent := i.Config; parent != nil {
		inherited := ManifestSettings{}
		if opts.Inherit.CheckUpdates {
			inherited.CheckUpdates = parent.CheckUpdates
		}
		if opts.Inherit.RepoUpdate {
			inherited.RepoUpdate = parent.RepoUpdate
		}
		if opts.Inherit.Defaults {
			inherited.Defaults = parent.Defaults
		}
		if opts.Inherit.Env {
			inherited.Env = parent.Env
			inherited.PlatformEnv = parent.PlatformEnv
		}
		if opts.Inherit.MachineAliases {
			inherited.MachineAliases = parent.MachineAliases
		}
		logger.Debug("Manifest %s inherits %+v", *i.GetData().Name, opts.Inherit)
		applyManifestSettings(config, &inherited)
	}
	if opts.Overrides != nil {
		applyManifestSettings(config, opts.Overrides)
	}
	if config.CheckUpdates == nil {
		config.CheckUpdates = lo.ToPtr(false)
	}
	return config
}

// applyManifestSettings merges settings into config. Maps merge key by key and `defaults` merge
// type by type, with the values in settings winning. The maps are copied so that a manifest never
// writes into the config it inherited from.
func applyManifestSettings(config *appconfig.AppConfig, settings *ManifestSettings) {
	if settings.CheckUpdates != nil {
		config.CheckUpdates = lo.ToPtr(*settings.CheckUpdates)
	}
	config.RepoUpdate = mergeMap(config.RepoUpdate, settings.RepoUpdate)
	config.Env = mergeMap(config.Env, settings.Env)
	config.MachineAliases = mergeMap(config.MachineAliases, settings.MachineAliases)
	if settings.PlatformEnv != nil {
		if config.PlatformEnv == nil {
			config.PlatformEnv = &platform.PlatformMap[map[string]string]{}
		}
		config.PlatformEnv.MacOS = mergeMap(config.PlatformEnv.MacOS, settings.PlatformEnv.MacOS)
		config.PlatformEnv.Linux = mergeMap(config.PlatformEnv.Linux, settings.PlatformEnv.Linux)
		config.PlatformEnv.Windows = mergeMap(config.PlatformEnv.Windows, settings.PlatformEnv.Windows)
	}
	if settings.Defaults != nil && settings.Defaults.Type != nil {
		if config.Defaults == nil {
			config.Defaults = &appconfig.AppConfigDefaults{}
		}
		config.Defaults.Type = mergeMap(config.Defaults.Type, settings.Defaults.Type)
	}
}

// mergeMap returns a fresh map holding base with overlay applied on top, or base when there is
// nothing to apply.
func mergeMap[K comparable, V any](base, overlay *map[K]V) *map[K]V {
	if overlay == nil {
		return base
	}
	out := map[K]V{}
	if base != nil {
		maps.Copy(out, *base)
	}
	maps.Copy(out, *overlay)
	return &out
}

// scopeManifestEnv applies the loaded manifest's env to the process for the duration of its
// install. When env is not inherited, every variable the loading config set (including those its
// env_command loaded) first goes back to its value from before sofmani applied any config, so the
// manifest cannot read it from the process either.
func (i *ManifestInstaller) scopeManifestEnv() (func(), error) {
	var reset []string
	if i.Config != nil && !i.GetOpts().Inherit.Env {
		parentEnv := utils.CombineEnvMaps(i.Config.Env, i.Config.PlatformEnv.Resolve(), &i.Config.LoadedEnv)
		reset = slices.Collect(maps.Keys(parentEnv))
	}
	config := i.ManifestConfig
	set := utils.ResolveEnvMap(utils.CombineEnvMaps(config.Env, config.PlatformEnv.Resolve()))
	return utils.ScopeEnv(set, reset)
}

func NewManifestInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *ManifestInstaller {
	return &ManifestInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Info:          installer,
	}
}
