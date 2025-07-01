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
	cmd := "brew install"
	if i.IsCask() {
		cmd += " --cask"
	}
	return i.RunCmdAsFile(fmt.Sprintf("%s %s", cmd, name))
}

// Update implements IInstaller.
func (i *BrewInstaller) Update() error {
	name := i.GetFullName()
	cmd := "brew upgrade"
	if i.IsCask() {
		cmd += " --cask"
	}
	return i.RunCmdAsFile(fmt.Sprintf("%s %s", cmd, name))
}

// GetFullName returns the full name of the package, including the tap if specified.
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
	cmd := exec.Command("brew", "outdated", "--json", name)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("Failed to get stdout pipe for brew command, error: %v", err)
		return false, fmt.Errorf("failed to get stdout: %w", err)
	}
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		logger.Error("Failed to start brew command, error: %v", err)
		return false, fmt.Errorf("failed to start brew: %w", err)
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
			fmt.Fprintln(logSink, line)
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
	return i.RunCmdGetSuccess(utils.GetShellWhich(), i.GetBinName())
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
