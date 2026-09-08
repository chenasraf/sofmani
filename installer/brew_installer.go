package installer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/utils"
)

// BrewInstaller is an installer for Homebrew packages.
type BrewInstaller struct {
	InstallerBase
	// Config is the application configuration.
	Config *appconfig.AppConfig
	// Info is the installer data.
	Info *appconfig.InstallerData
}

// BrewOpts represents options for the BrewInstaller.
type BrewOpts struct {
	// Tap is the Homebrew tap to use for the package.
	Tap *string
	// Cask installs the formula as a cask instead of a regular package.
	Cask *bool
	// Flags is a string of additional flags to pass to the brew command.
	Flags *string
	// InstallFlags is a string of additional flags to pass only during install.
	InstallFlags *string
	// UpdateFlags is a string of additional flags to pass only during update.
	UpdateFlags *string
}

// Validate validates the installer configuration.
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
	opts := i.GetOpts()
	disableBrewAsk()
	if err := i.ensureTapped(); err != nil {
		return err
	}
	i.handleBrewRepoUpdate()
	cmd := "brew install"
	if i.IsVerbose() {
		cmd += " --verbose"
	}
	if i.IsCask() {
		cmd += " --cask"
	}
	if opts.InstallFlags != nil {
		cmd += " " + *opts.InstallFlags
	} else if opts.Flags != nil {
		cmd += " " + *opts.Flags
	}
	err := i.RunCmdAsFile(fmt.Sprintf("%s %s", cmd, name))
	i.markBrewRepoUpdated()
	return err
}

// Update implements IInstaller.
func (i *BrewInstaller) Update() error {
	name := i.GetFullName()
	opts := i.GetOpts()
	disableBrewAsk()
	if err := i.ensureTapped(); err != nil {
		return err
	}
	i.handleBrewRepoUpdate()
	cmd := "brew upgrade"
	if i.IsVerbose() {
		cmd += " --verbose"
	}
	if i.IsCask() {
		cmd += " --cask"
	}
	if opts.UpdateFlags != nil {
		cmd += " " + *opts.UpdateFlags
	} else if opts.Flags != nil {
		cmd += " " + *opts.Flags
	}
	err := i.RunCmdAsFile(fmt.Sprintf("%s %s", cmd, name))
	i.markBrewRepoUpdated()
	return err
}

// GetFullName returns the full name of the package, including the tap if specified.
func (i *BrewInstaller) GetFullName() string {
	name := *i.Info.Name
	if i.GetOpts().Tap != nil {
		name = *i.GetOpts().Tap + "/" + name
	}
	return name
}

// disableBrewAsk suppresses Homebrew's interactive "Ask mode" confirmation prompt
// (the default since Homebrew 6.x) so installs and upgrades run non-interactively.
func disableBrewAsk() {
	_ = os.Setenv("HOMEBREW_NO_ASK", "1")
}

// ensureTapped runs `brew tap <tap>` followed by `brew trust --tap <tap>`, once per process
// for the installer's configured tap. Homebrew requires taps to be added explicitly before
// installing from them, and refuses to load a formula or cask from a non-official tap that
// has not been trusted, unless HOMEBREW_NO_REQUIRE_TAP_TRUST is set.
func (i *BrewInstaller) ensureTapped() error {
	opts := i.GetOpts()
	if opts.Tap == nil {
		return nil
	}
	tap := *opts.Tap
	return RunRepoUpdateOnce("brew-tap:"+tap, func() error {
		logger.Debug("Tapping brew tap %s", tap)
		if err := runBrewCmd("tap", tap); err != nil {
			return fmt.Errorf("failed to tap %s: %w", tap, err)
		}
		if os.Getenv("HOMEBREW_NO_REQUIRE_TAP_TRUST") != "" {
			return nil
		}
		logger.Debug("Trusting brew tap %s", tap)
		if err := runBrewCmd("trust", "--tap", tap); err != nil {
			// Homebrew versions without `brew trust` don't require trust to begin with, so
			// let the install itself report the problem if the tap really is untrusted.
			logger.Warn("Failed to trust tap %s: %v", tap, err)
		}
		return nil
	})
}

// runBrewCmd runs a brew command attached to the current process' stdio.
func runBrewCmd(args ...string) error {
	cmd := exec.Command("brew", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// handleBrewRepoUpdate manages brew's auto-update behavior according to the configured mode.
// For "once" (default): lets the first brew command auto-update, then suppresses subsequent ones.
// For "always": does nothing (brew auto-updates every time).
// For "never": always suppresses auto-update.
func (i *BrewInstaller) handleBrewRepoUpdate() {
	mode := i.Config.GetRepoUpdateMode(appconfig.InstallerTypeBrew)
	switch mode {
	case appconfig.RepoUpdateNever:
		_ = os.Setenv("HOMEBREW_NO_AUTO_UPDATE", "1")
	case appconfig.RepoUpdateAlways:
		// Let brew auto-update every time
	default: // once
		if IsRepoUpdated("brew") {
			_ = os.Setenv("HOMEBREW_NO_AUTO_UPDATE", "1")
		}
	}
}

// markBrewRepoUpdated marks brew's repo as updated (for "once" mode tracking).
func (i *BrewInstaller) markBrewRepoUpdated() {
	if i.Config.GetRepoUpdateMode(appconfig.InstallerTypeBrew) == appconfig.RepoUpdateOnce {
		MarkRepoUpdated("brew")
	}
}

// CheckNeedsUpdate implements IInstaller.
func (i *BrewInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}

	if err := i.ensureTapped(); err != nil {
		return false, err
	}
	i.handleBrewRepoUpdate()
	name := i.GetFullName()
	cmd := exec.Command("brew", "outdated", "--json", name)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("Failed to get stdout pipe for brew command, error: %v", err)
		return false, fmt.Errorf("failed to get stdout for `brew outdated %s`: %w", name, err)
	}
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		logger.Error("Failed to start brew command, error: %v", err)
		return false, fmt.Errorf("failed to start `brew outdated %s`: %w", name, err)
	}

	updateNeeded, parseErr := parseBrewOutdatedOutput(stdoutPipe, os.Stdout)

	waitErr := cmd.Wait()
	if waitErr != nil {
		exitErr, ok := waitErr.(*exec.ExitError)
		if ok {
			exitCode := exitErr.ExitCode()
			// 0 = no update, 1 = update available → both acceptable
			if exitCode != 0 && exitCode != 1 {
				logger.Error("Brew command failed with unexpected code %d", exitCode)
				return false, waitErr
			}
		} else {
			// Non-exit error (e.g. I/O), return as-is
			logger.Error("Brew command failed, non-exit error: %v", waitErr)
			return false, waitErr
		}
	}

	if parseErr != nil {
		logger.Error("Failed to parse brew output, error: %v", parseErr)
		return false, fmt.Errorf("failed to parse brew output: %w", parseErr)
	}

	i.markBrewRepoUpdated()
	return updateNeeded, nil
}

// parseBrewOutdatedOutput parses the JSON output of `brew outdated --json`.
// It returns true if an update is needed, false otherwise.
func parseBrewOutdatedOutput(input io.Reader, logSink io.Writer) (bool, error) {
	var jsonBuf bytes.Buffer
	scanner := bufio.NewScanner(input)
	inJSON := false

	logger.Debug("Parsing brew outdated output")
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "{") {
			inJSON = true
		}

		if inJSON {
			jsonBuf.WriteString(line + "\n")
		} else {
			_, err := fmt.Fprintln(logSink, line)
			if err != nil {
				logger.Error("Failed to write line to log sink, error: %v", err)
				return false, fmt.Errorf("failed to write line to log sink: %w", err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return false, err
	}

	// Parse JSON
	type brewOutdatedJSON struct {
		Formulae []any `json:"formulae"`
		Casks    []any `json:"casks"`
	}
	var parsed brewOutdatedJSON
	logger.Debug("Unmarshalling JSON from brew outdated output: %s", jsonBuf.String())
	if err := json.Unmarshal(jsonBuf.Bytes(), &parsed); err != nil {
		logger.Error("Failed to unmarshal JSON from brew outdated output, error: %v", err)
		return false, err
	}
	return len(parsed.Formulae) > 0 || len(parsed.Casks) > 0, nil
}

// CheckIsInstalled implements IInstaller.
func (i *BrewInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	// Ask Homebrew rather than looking the binary up on PATH: casks often ship only an
	// .app bundle, and formulae may install a binary under a different name.
	for _, scope := range []string{brewListAll, brewListCasks} {
		installed, err := brewIsPackageInstalled(scope, *i.GetData().Name)
		if err != nil {
			return false, err
		}
		if installed {
			return true, nil
		}
	}
	return false, nil
}

const (
	brewListAll   = "all"
	brewListCasks = "casks"
)

var (
	brewListMu    sync.Mutex
	brewListCache = map[string]map[string]bool{}
)

// brewIsPackageInstalled reports whether a package appears in the given `brew list` scope.
func brewIsPackageInstalled(scope string, name string) (bool, error) {
	names, err := brewInstalledNames(scope)
	if err != nil {
		return false, err
	}
	// A tap-qualified name is not what `brew list` reports, so compare on the bare name.
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	return names[name], nil
}

// brewInstalledNames returns the names of the installed packages in the given `brew list`
// scope. `brew list` covers both formulae and casks, and the cask-only scope catches
// Homebrew versions that omit casks from it. Each scope is fetched once per run, since one
// `brew list` costs about as much as querying a single package.
func brewInstalledNames(scope string) (map[string]bool, error) {
	brewListMu.Lock()
	defer brewListMu.Unlock()
	if names, ok := brewListCache[scope]; ok {
		return names, nil
	}
	args := []string{"list", "--versions"}
	if scope == brewListCasks {
		args = []string{"list", "--cask", "--versions"}
	}
	logger.Debug("Listing installed brew packages (%s)", scope)
	out, err := utils.RunCmdGetOutput(nil, "brew", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list installed brew packages: %w", err)
	}
	names := parseBrewListOutput(out)
	brewListCache[scope] = names
	return names, nil
}

// parseBrewListOutput parses the output of `brew list --versions` into a set of package names.
// Each line holds a name followed by the installed versions.
func parseBrewListOutput(out []byte) map[string]bool {
	names := map[string]bool{}
	for line := range strings.SplitSeq(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		names[fields[0]] = true
	}
	return names
}

// ResetBrewListCache clears the cached `brew list` output. Intended for testing.
func ResetBrewListCache() {
	brewListMu.Lock()
	defer brewListMu.Unlock()
	brewListCache = map[string]map[string]bool{}
}

// GetData implements IInstaller.
func (i *BrewInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

// GetOpts returns the parsed options for the BrewInstaller.
func (i *BrewInstaller) GetOpts() *BrewOpts {
	opts := &BrewOpts{}
	info := i.Info
	if info.Opts != nil {
		if tap, ok := (*info.Opts)["tap"].(string); ok {
			opts.Tap = &tap
		}
		if caskVal, ok := (*info.Opts)["cask"].(bool); ok {
			opts.Cask = &caskVal
		}
		if flags, ok := (*info.Opts)["flags"].(string); ok {
			opts.Flags = &flags
		}
		if installFlags, ok := (*info.Opts)["install_flags"].(string); ok {
			opts.InstallFlags = &installFlags
		}
		if updateFlags, ok := (*info.Opts)["update_flags"].(string); ok {
			opts.UpdateFlags = &updateFlags
		}
	}
	return opts
}

func (i *BrewInstaller) IsCask() bool {
	opts := i.GetOpts()
	return opts.Cask != nil && *opts.Cask
}

// GetBinName returns the binary name for the installer.
// It uses the BinName from the installer data if provided, otherwise it uses the installer name.
func (i *BrewInstaller) GetBinName() string {
	info := i.GetData()
	if info.BinName != nil && len(*info.BinName) > 0 {
		return *info.BinName
	}
	return *info.Name
}

// NewBrewInstaller creates a new BrewInstaller.
func NewBrewInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *BrewInstaller {
	i := &BrewInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Info:          installer,
	}

	return i
}
