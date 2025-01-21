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
	tmpOut, err := os.Create(fmt.Sprintf("%s/%s.download", tmpDir, name))
	defer tmpOut.Close()

	err = os.MkdirAll(*opts.Destination, 0755)
	if err != nil {
		return err
	}
	// defer os.RemoveAll(tmpDir)

	tag, err := i.GetLatestTag()
	if err != nil {
		return err
	}

	version := tag
	if strings.HasPrefix(tag, "v") {
		version = strings.TrimPrefix(tag, "v")
	}
	filename := i.GetFilename()
	filename = strings.ReplaceAll(filename, "{tag}", tag)
	filename = strings.ReplaceAll(filename, "{version}", version)
	if filename == "" {
		return fmt.Errorf("No download filename provided")
	}
	downloadUrl := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", *opts.Repository, tag, filename)
	logger.Debug("Downloading from %s", downloadUrl)
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

	logger.Debug("Strategy %s", strategy)

	success := false

	logger.Debug("Creating file %s", fmt.Sprintf("%s/%s", *opts.Destination, i.GetBinName()))
	out, err := os.Create(fmt.Sprintf("%s/%s", *opts.Destination, i.GetBinName()))
	defer out.Close()
	if err != nil {
		return err
	}

	switch strategy {
	case GitHubReleaseInstallStrategyTar:
		logger.Debug("Extracting tar file %s", tmpOut.Name())
		success, err = i.RunCmdGetSuccess("tar", "-xvf", tmpOut.Name(), "-C", tmpDir)
		if err != nil {
			return err
		}
		success, err = i.CopyExtractedFile(out, tmpDir)
		if err != nil {
			return err
		}
	case GitHubReleaseInstallStrategyZip:
		logger.Debug("Extracting zip file %s", tmpOut.Name())
		success, err = i.RunCmdGetSuccess("unzip", tmpOut.Name(), "-d", tmpDir)
		if err != nil {
			return err
		}
		success, err = i.CopyExtractedFile(out, tmpDir)
		if err != nil {
			return err
		}
	default:
		io.Copy(out, tmpOut)
		success = true
		err = nil
	}

	if !success || err != nil {
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

// CheckIsInstalled implements IInstaller.
func (i *GitHubReleaseInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}
	logger.Debug("Checking if %s is installed on %s", *i.Data.Name, filepath.Join(i.GetInstallDir(), *i.Data.Name))
	return utils.PathExists(filepath.Join(i.GetInstallDir(), *i.Data.Name))
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

func (i *GitHubReleaseInstaller) GetBinName() string {
	if i.Data.BinName != nil {
		return *i.Data.BinName
	}
	return filepath.Base(*i.Data.Name)
}

func (i *GitHubReleaseInstaller) CopyExtractedFile(out *os.File, tmpDir string) (bool, error) {
	binFile, err := os.Create(out.Name())
	defer binFile.Close()
	if err != nil {
		return false, err
	}
	tmpBinFile, err := os.Open(filepath.Join(tmpDir, i.GetBinName()))
	logger.Debug("Copying file %s to %s", tmpBinFile.Name(), binFile.Name())

	n, err := io.Copy(binFile, tmpBinFile)
	if err != nil {
		return false, err
	}
	if n == 0 {
		return false, fmt.Errorf("No data was written to the file")
	}
	return true, nil
}

func (i *GitHubReleaseInstaller) GetCachedTag() (string, error) {
	logger.Debug("Getting cached tag for %s", *i.Data.Name)
	cacheDir, err := utils.GetCacheDir()
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
	logger.Debug("Got cached tag %s for %s", strings.TrimSpace(string(contents)), *i.Data.Name)
	return strings.TrimSpace(string(contents)), nil
}

func (i *GitHubReleaseInstaller) UpdateCache(tag string) error {
	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return err
	}
	cacheFile := fmt.Sprintf("%s/%s", cacheDir, *i.Data.Name)
	logger.Debug("Updating cache file %s with %s", cacheFile, tag)
	err = os.WriteFile(cacheFile, []byte(tag), 0644)
	if err != nil {
		return err
	}
	return nil
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
	return opts
}

func (i *GitHubReleaseInstaller) GetLatestTag() (string, error) {
	latestReleaseUrl := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", *i.GetOpts().Repository)
	logger.Debug("Getting latest release from %s", latestReleaseUrl)
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
	logger.Debug("Latest release is %s", tag)
	return tag, nil
}

func (i *GitHubReleaseInstaller) GetFilename() string {
	opts := i.GetOpts()
	if opts.DownloadFilename != nil {
		return *opts.DownloadFilename.Resolve()
	}
	return ""
}

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

func (i *GitHubReleaseInstaller) GetInstallDir() string {
	return i.GetDestination()
}

func NewGitHubReleaseInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *GitHubReleaseInstaller {
	i := &GitHubReleaseInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Data:          installer,
	}

	return i
}
