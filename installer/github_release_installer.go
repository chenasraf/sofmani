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

type GitHubReleaseInstaller struct {
	InstallerBase
	Config *appconfig.AppConfig
	Data   *appconfig.InstallerData
}

type GitHubReleaseOpts struct {
	Repository       *string
	Destination      *string
	DownloadFilename *platform.PlatformMap[string]
	Strategy         *GitHubReleaseInstallStrategy
}

type GitHubReleaseInstallStrategy string

const (
	GitHubReleaseInstallStrategyNone GitHubReleaseInstallStrategy = "none"
	GitHubReleaseInstallStrategyTar  GitHubReleaseInstallStrategy = "tar"
	GitHubReleaseInstallStrategyZip  GitHubReleaseInstallStrategy = "zip"
)

// Install implements IInstaller.
func (i *GitHubReleaseInstaller) Install() error {
	opts := i.GetOpts()
	data := i.GetData()
	name := *data.Name
	tmpDir, err := os.MkdirTemp("", "sofmani")
	if err != nil {
		return err
	}
	tmpOut, err := os.Create(fmt.Sprintf("%s/%s", tmpDir, name))
	logger.Debug("tmpOut: %v", tmpOut)

	err = os.MkdirAll(*opts.Destination, 0755)
	if err != nil {
		return err
	}

	out, err := os.Create(fmt.Sprintf("%s/%s", *opts.Destination, name))
	logger.Debug("out: %v, %v", out, err)
	defer out.Close()
	if err != nil {
		return err
	}

	tag, err := i.GetLatestTag()
	if err != nil {
		return err
	}

	replTag := tag
	if strings.HasPrefix(tag, "v") {
		replTag = strings.TrimPrefix(tag, "v")
	}
	filename := i.GetFilename()
	filename = strings.ReplaceAll(filename, "{tag}", replTag)
	filename = strings.ReplaceAll(filename, "{version}", tag)
	if filename == "" {
		return fmt.Errorf("No download filename provided")
	}
	downloadUrl := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", *opts.Repository, tag, filename)
	logger.Debug("Downloading %s", downloadUrl)
	resp, err := http.Get(downloadUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	n, err := io.Copy(tmpOut, resp.Body)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("No data was written to the file")
	}

	strategy := GitHubReleaseInstallStrategyNone

	if opts.Strategy != nil {
		strategy = *opts.Strategy
	}

	success := false

	switch strategy {
	case GitHubReleaseInstallStrategyTar:
		success, err = i.RunCmdGetSuccess("tar", "-xvf", tmpOut.Name(), "-C", *opts.Destination)
	case GitHubReleaseInstallStrategyZip:
		success, err = i.RunCmdGetSuccess("unzip", tmpOut.Name(), "-d", *opts.Destination)
	default:
		io.Copy(out, tmpOut)
		success = true
		err = nil
	}

	if !success {
		return errors.Join(fmt.Errorf("Failed to extract the downloaded file"), err)
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
	if latest != strings.TrimSpace(latest) {
		return true, nil
	}
	return false, nil
}

func (i *GitHubReleaseInstaller) GetCachedTag() (string, error) {
	cacheDir, err := utils.GetCacheDir()
	logger.Debug("cacheDir: %v", cacheDir)
	if err != nil {
		return "", err
	}
	cacheFile := fmt.Sprintf("%s/%s", cacheDir, *i.Data.Name)
	exists, err := utils.PathExists(cacheFile)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", nil
	}
	reader, err := os.Open(cacheFile)
	contents, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(contents)), nil
}

func (i *GitHubReleaseInstaller) UpdateCache(tag string) error {
	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return err
	}
	cacheFile := fmt.Sprintf("%s/%s", cacheDir, *i.Data.Name)
	err = os.WriteFile(cacheFile, []byte(tag), 0644)
	if err != nil {
		return err
	}
	return nil
}

// CheckIsInstalled implements IInstaller.
func (i *GitHubReleaseInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	return utils.PathExists(i.GetInstallDir())
}

// GetData implements IInstaller.
func (i *GitHubReleaseInstaller) GetData() *appconfig.InstallerData {
	return i.Data
}

func (i *GitHubReleaseInstaller) GetOpts() *GitHubReleaseOpts {
	opts := &GitHubReleaseOpts{}
	info := i.Data
	if info.Opts != nil {
		if repository, ok := (*info.Opts)["repository"].(string); ok {
			repository = utils.GetRealPath(i.GetData().Environ(), repository)
			opts.Repository = &repository
		}
		if destination, ok := (*info.Opts)["destination"].(string); ok {
			destination = utils.GetRealPath(i.GetData().Environ(), destination)
			opts.Destination = &destination
		}
		if filename, ok := (*info.Opts)["download_filename"].(string); ok {
			opts.DownloadFilename = &platform.PlatformMap[string]{
				MacOS:   &filename,
				Linux:   &filename,
				Windows: &filename,
			}
		} else if filenameMap, ok := (*info.Opts)["download_filename"].(map[string]*string); ok {
			opts.DownloadFilename = &platform.PlatformMap[string]{
				MacOS:   filenameMap["macos"],
				Linux:   filenameMap["linux"],
				Windows: filenameMap["windows"],
			}
		}
		if strategy, ok := (*info.Opts)["strategy"].(string); ok {
			strat := GitHubReleaseInstallStrategy(strings.ToLower(strategy))
			opts.Strategy = &strat
		}
	}
	logger.Debug("GitHubReleaseInstaller.GetOpts: %v", opts.DownloadFilename)
	return opts
}

func (i *GitHubReleaseInstaller) GetLatestTag() (string, error) {
	latestReleaseUrl := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", *i.GetOpts().Repository)
	resp, err := http.Get(latestReleaseUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	jsonMap := make(map[string]any)
	err = json.Unmarshal(contents, &jsonMap)
	if err != nil {
		return "", err
	}
	tag := jsonMap["tag_name"].(string)

	return tag, nil
}

func (i *GitHubReleaseInstaller) GetFilename() string {
	opts := i.GetOpts()
	if opts.PlatformDownloadFilenames != nil {
		filename := *opts.DownloadFilename
		return opts.PlatformDownloadFilenames.ResolveWithFallback(platform.PlatformMap[string]{
			MacOS:   &filename,
			Linux:   &filename,
			Windows: &filename,
		})
	}
	if opts.DownloadFilename != nil {
		return *opts.DownloadFilename
	}
	return ""
}

func (i *GitHubReleaseInstaller) GetDestination() string {
	if i.GetOpts().Repository != nil {
		return *i.GetOpts().Repository
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}

func (i *GitHubReleaseInstaller) GetInstallDir() string {
	return filepath.Join(i.GetDestination(), filepath.Base(*i.Data.Name))
}

func NewGitHubReleaseInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *GitHubReleaseInstaller {
	i := &GitHubReleaseInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Data:          installer,
	}

	return i
}
