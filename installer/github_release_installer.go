package installer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
	"github.com/chenasraf/sofmani/utils"
)

// GitHubReleaseInstaller is an installer for GitHub releases.
type GitHubReleaseInstaller struct {
	InstallerBase
	// Config is the application configuration.
	Config *appconfig.AppConfig
	// Info is the installer data.
	Info *appconfig.InstallerData
}

// GitHubReleaseOpts represents options for the GitHubReleaseInstaller.
type GitHubReleaseOpts struct {
	// Repository is the GitHub repository (e.g., "owner/repo").
	Repository *string
	// Destination is the directory where the release asset will be installed.
	Destination *string
	// DownloadFilename is a platform-specific map of the filename to download from the release.
	// Supports Go template syntax with variables: {{ .Tag }}, {{ .Version }}, {{ .Arch }}, {{ .ArchAlias }}, {{ .ArchGnu }}, {{ .OS }}.
	// Legacy placeholders {tag}, {version}, {arch}, {arch_alias}, {arch_gnu}, {os} are deprecated but still supported.
	DownloadFilename *platform.PlatformMap[string]
	// Strategy is the installation strategy to use (none, tar, zip).
	Strategy *GitHubReleaseInstallStrategy
	// GithubToken is the GitHub personal access token for authenticated API requests.
	// Supports environment variable expansion (e.g., "$GITHUB_TOKEN" or "${GITHUB_TOKEN}").
	GithubToken *string
}

// GitHubReleaseInstallStrategy represents the installation strategy for a GitHub release.
type GitHubReleaseInstallStrategy string

// Constants for GitHub release installation strategies.
const (
	GitHubReleaseInstallStrategyNone GitHubReleaseInstallStrategy = "none" // GitHubReleaseInstallStrategyNone means no special handling, just download the file.
	GitHubReleaseInstallStrategyTar  GitHubReleaseInstallStrategy = "tar"  // GitHubReleaseInstallStrategyTar means extract a tar archive.
	GitHubReleaseInstallStrategyZip  GitHubReleaseInstallStrategy = "zip"  // GitHubReleaseInstallStrategyZip means extract a zip archive.
)

// Validate validates the installer configuration.
func (i *GitHubReleaseInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	info := i.GetData()
	opts := i.GetOpts()
	if opts.Repository == nil || len(*opts.Repository) == 0 {
		errors = append(errors, ValidationError{FieldName: "repository", Message: validationIsRequired(), InstallerName: *info.Name})
	}
	if opts.Destination == nil || len(*opts.Destination) == 0 {
		errors = append(errors, ValidationError{FieldName: "destination", Message: validationIsRequired(), InstallerName: *info.Name})
	}
	if opts.DownloadFilename == nil || len(*opts.DownloadFilename.Resolve()) == 0 {
		errors = append(errors, ValidationError{FieldName: "download_filename", Message: validationIsRequired(), InstallerName: *info.Name})
	} else if (*opts.DownloadFilename).Resolve() == nil || len(*(*opts.DownloadFilename).Resolve()) == 0 {
		errors = append(errors, ValidationError{FieldName: fmt.Sprintf("download_filename.%s", platform.GetPlatform()), Message: validationIsRequired(), InstallerName: *info.Name})
	}
	if opts.Strategy != nil {
		if *opts.Strategy != GitHubReleaseInstallStrategyNone && *opts.Strategy != GitHubReleaseInstallStrategyTar && *opts.Strategy != GitHubReleaseInstallStrategyZip {
			errors = append(errors, ValidationError{FieldName: "strategy", Message: validationInvalidFormat(), InstallerName: *info.Name})
		}
	}
	return errors
}

// Install implements IInstaller.
func (i *GitHubReleaseInstaller) Install() error {
	opts := i.GetOpts()
	data := i.GetData()
	name := *data.Name
	tmpDir, err := os.MkdirTemp("", "sofmani")
	if err != nil {
		return err
	}
	tmpOut, err := os.Create(fmt.Sprintf("%s/%s.download", tmpDir, name))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		if cerr := tmpOut.Close(); cerr != nil {
			logger.Warn("failed to close tmpOut file: %v", cerr)
		}
	}()

	err = os.MkdirAll(*opts.Destination, 0755)
	if err != nil {
		return err
	}
	// defer os.RemoveAll(tmpDir)

	tag, err := i.GetLatestTag()
	if err != nil {
		return err
	}

	filename := i.GetFilename()
	if filename == "" {
		return fmt.Errorf("no download filename provided")
	}
	templateVars := NewTemplateVars(tag)
	filename, err = ApplyTemplate(filename, templateVars, name)
	if err != nil {
		return fmt.Errorf("failed to apply template to filename: %w", err)
	}
	downloadUrl := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", *opts.Repository, tag, filename)
	logger.Debug("Downloading from %s", downloadUrl)

	req, err := http.NewRequest("GET", downloadUrl, nil)
	if err != nil {
		return err
	}
	if opts.GithubToken != nil && *opts.GithubToken != "" {
		req.Header.Set("Authorization", "Bearer "+*opts.GithubToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logger.Warn("failed to close response body: %v", cerr)
		}
	}()

	n, err := io.Copy(tmpOut, resp.Body)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("no data was written to the file")
	}

	strategy := GitHubReleaseInstallStrategyNone

	if opts.Strategy != nil {
		strategy = *opts.Strategy
	}

	logger.Debug("Strategy %s", strategy)

	success := false

	outPath := filepath.Join(*opts.Destination, i.GetBinName())
	logger.Debug("Creating file %s", outPath)

	// Remove existing file first to avoid "text file busy" error on Linux
	// when updating a running executable
	if err := os.Remove(outPath); err != nil && !os.IsNotExist(err) {
		logger.Debug("Could not remove existing file: %v", err)
	}

	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if cerr := out.Close(); cerr != nil {
			logger.Warn("failed to close output file: %v", cerr)
		}
	}()

	switch strategy {
	case GitHubReleaseInstallStrategyTar:
		logger.Debug("Extracting tar file %s", tmpOut.Name())
		success, err = i.RunCmdGetSuccess("tar", "-xvf", tmpOut.Name(), "-C", tmpDir)
		if !success {
			return fmt.Errorf("failed to extract tar file: %w", err)
		}
		if err != nil {
			return err
		}
		success, err = i.CopyExtractedFile(out, tmpDir)
		if !success {
			return fmt.Errorf("failed to copy extracted file: %w", err)
		}
		if err != nil {
			return err
		}
	case GitHubReleaseInstallStrategyZip:
		logger.Debug("Extracting zip file %s", tmpOut.Name())
		success, err = i.RunCmdGetSuccess("unzip", tmpOut.Name(), "-d", tmpDir)
		if !success {
			return fmt.Errorf("failed to extract zip file: %w", err)
		}
		if err != nil {
			return err
		}
		success, err = i.CopyExtractedFile(out, tmpDir)
		if !success {
			return fmt.Errorf("failed to copy extracted file: %w", err)
		}
		if err != nil {
			return err
		}
	default:
		// Seek back to beginning of temp file before copying
		if _, err = tmpOut.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek temp file: %w", err)
		}
		_, err = io.Copy(out, tmpOut)
		if err != nil {
			return fmt.Errorf("failed to copy downloaded file to output: %w", err)
		}
		success = true
		err = nil
	}

	if !success {
		return fmt.Errorf("failed to copy the downloaded file to the output file")
	}
	if err != nil {
		return errors.Join(fmt.Errorf("failed to extract the downloaded file"), err)
	}

	// Make the file executable
	if err = os.Chmod(outPath, 0755); err != nil {
		return fmt.Errorf("failed to make file executable: %w", err)
	}

	err = i.UpdateCache(tag)
	if err != nil {
		return err
	}

	return nil
}

// Update implements IInstaller.
func (i *GitHubReleaseInstaller) Update() error {
	return i.Install()
}

// CheckIsInstalled implements IInstaller.
func (i *GitHubReleaseInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	logger.Debug("Checking if %s is installed on %s", *i.Info.Name, filepath.Join(i.GetInstallDir(), *i.Info.Name))
	return utils.PathExists(filepath.Join(i.GetInstallDir(), *i.Info.Name))
}

// CheckNeedsUpdate implements IInstaller.
func (i *GitHubReleaseInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}
	cachedTag, err := i.GetCachedTag()
	if err != nil {
		return false, err
	}
	if cachedTag == "" {
		return true, nil
	}
	latest, err := i.GetLatestTag()
	if err != nil {
		return false, err
	}
	if latest != cachedTag {
		return true, nil
	}
	return false, nil
}

// GetBinName returns the binary name for the installer.
// It uses the BinName from the installer data if provided, otherwise it uses the base name of the installer name.
func (i *GitHubReleaseInstaller) GetBinName() string {
	if i.Info.BinName != nil {
		return *i.Info.BinName
	}
	return filepath.Base(*i.Info.Name)
}

// CopyExtractedFile copies the extracted file from a temporary directory to the final destination.
func (i *GitHubReleaseInstaller) CopyExtractedFile(out *os.File, tmpDir string) (bool, error) {
	binFile, err := os.Create(out.Name())
	if err != nil {
		return false, fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if cerr := binFile.Close(); cerr != nil {
			logger.Warn("failed to close binFile %s: %v", binFile.Name(), cerr)
		}
	}()
	tmpBinFile, err := os.Open(filepath.Join(tmpDir, i.GetBinName()))
	if err != nil {
		return false, fmt.Errorf("failed to open temporary file: %w", err)
	}
	logger.Debug("Copying file %s to %s", tmpBinFile.Name(), binFile.Name())

	n, err := io.Copy(binFile, tmpBinFile)
	if err != nil {
		return false, err
	}
	if n == 0 {
		return false, fmt.Errorf("no data was written to the file")
	}
	return true, nil
}

// GetCachedTag retrieves the cached tag for the release from the cache directory.
func (i *GitHubReleaseInstaller) GetCachedTag() (string, error) {
	logger.Debug("Getting cached tag for %s", *i.Info.Name)
	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return "", err
	}
	cacheFile := fmt.Sprintf("%s/%s", cacheDir, *i.Info.Name)
	exists, err := utils.PathExists(cacheFile)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", nil
	}
	reader, err := os.Open(cacheFile)
	if err != nil {
		return "", fmt.Errorf("failed to open cache file %s: %w", cacheFile, err)
	}
	contents, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read cache file %s: %w", cacheFile, err)
	}
	logger.Debug("Got cached tag %s for %s", strings.TrimSpace(string(contents)), *i.Info.Name)
	return strings.TrimSpace(string(contents)), nil
}

// UpdateCache updates the cached tag for the release in the cache directory.
func (i *GitHubReleaseInstaller) UpdateCache(tag string) error {
	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return err
	}
	cacheFile := fmt.Sprintf("%s/%s", cacheDir, *i.Info.Name)
	logger.Debug("Updating cache file %s with %s", cacheFile, tag)
	err = os.WriteFile(cacheFile, []byte(tag), 0644)
	if err != nil {
		return err
	}
	return nil
}

// GetData implements IInstaller.
func (i *GitHubReleaseInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

// GetOpts returns the parsed options for the GitHubReleaseInstaller.
func (i *GitHubReleaseInstaller) GetOpts() *GitHubReleaseOpts {
	opts := &GitHubReleaseOpts{}
	info := i.Info
	if info.Opts != nil {
		if repository, ok := (*info.Opts)["repository"].(string); ok {
			repository = utils.GetRealPath(i.GetData().Environ(), repository)
			opts.Repository = &repository
		}
		if destination, ok := (*info.Opts)["destination"].(string); ok {
			destination = utils.GetRealPath(i.GetData().Environ(), destination)
			opts.Destination = &destination
		}
		if filename, ok := (*info.Opts)["download_filename"]; ok {
			opts.DownloadFilename = platform.NewPlatformMap[string](filename)
		}
		if strategy, ok := (*info.Opts)["strategy"].(string); ok {
			strat := GitHubReleaseInstallStrategy(strings.ToLower(strategy))
			opts.Strategy = &strat
		}
		if token, ok := (*info.Opts)["github_token"].(string); ok {
			token = utils.GetRealPath(i.GetData().Environ(), token)
			opts.GithubToken = &token
		}
	}
	return opts
}

func (i *GitHubReleaseInstaller) GetLatestTag() (string, error) {
	opts := i.GetOpts()
	latestReleaseUrl := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", *opts.Repository)
	logger.Debug("Getting latest release from %s", latestReleaseUrl)

	req, err := http.NewRequest("GET", latestReleaseUrl, nil)
	if err != nil {
		return "", err
	}
	if opts.GithubToken != nil && *opts.GithubToken != "" {
		req.Header.Set("Authorization", "Bearer "+*opts.GithubToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Warn("Failed to close response body: %v", err)
		}
	}()
	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	jsonMap := make(map[string]any)
	err = json.Unmarshal(contents, &jsonMap)
	if err != nil {
		return "", err
	}
	tag, ok := jsonMap["tag_name"].(string)
	if !ok || tag == "" {
		logger.Warn("Invalid GitHub API response: %s", string(contents))
		if msg, ok := jsonMap["message"].(string); ok {
			return "", fmt.Errorf("GitHub API error: %s", msg)
		}
		return "", fmt.Errorf("no releases found for repository")
	}
	logger.Debug("Latest release is %s", tag)
	return tag, nil
}

// GetFilename returns the filename to download from the release, resolved for the current platform.
func (i *GitHubReleaseInstaller) GetFilename() string {
	opts := i.GetOpts()
	if opts.DownloadFilename != nil {
		return *opts.DownloadFilename.Resolve()
	}
	return ""
}

// GetDestination returns the destination directory for the release asset.
// It uses the Destination from the installer options if provided, otherwise it defaults to the current working directory.
func (i *GitHubReleaseInstaller) GetDestination() string {
	if i.GetOpts().Destination != nil {
		return *i.GetOpts().Destination
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}

// GetInstallDir returns the installation directory for the release asset.
// For GitHub releases, this is the same as the destination directory.
func (i *GitHubReleaseInstaller) GetInstallDir() string {
	return i.GetDestination()
}

// NewGitHubReleaseInstaller creates a new GitHubReleaseInstaller.
func NewGitHubReleaseInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *GitHubReleaseInstaller {
	i := &GitHubReleaseInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Info:          installer,
	}

	return i
}
