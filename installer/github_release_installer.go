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
	"github.com/chenasraf/sofmani/platform"
	"github.com/chenasraf/sofmani/utils"
)

type GitHubReleaseInstaller struct {
	InstallerBase
	Config *appconfig.AppConfig
	Info   *appconfig.InstallerData
}

type GitHubReleaseOpts struct {
	Repository                *string
	Destination               *string
	DownloadFilename          *string
	PlatformDownloadFilenames *platform.PlatformMap[string]
	Strategy                  *GitHubReleaseInstallStrategy
}

type GitHubReleaseInstallStrategy string

const (
	StrategyNone GitHubReleaseInstallStrategy = "none"
	StrategyTar  GitHubReleaseInstallStrategy = "tar"
	StrategyZip  GitHubReleaseInstallStrategy = "zip"
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

	out, err := os.Create(fmt.Sprintf("%s/%s", *opts.Destination, name))
	defer out.Close()
	if err != nil {
		return err
	}

	tag, err := i.GetLatestTag()
	if err != nil {
		return err
	}

	filename := strings.ReplaceAll(i.GetFilename(), "{tag}", tag)
	if filename == "" {
		return fmt.Errorf("No download filename provided")
	}
	downloadUrl := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", *opts.Repository, tag, filename)
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

	strategy := StrategyNone

	if opts.Strategy != nil {
		strategy = *opts.Strategy
	}

	success := false

	switch strategy {
	case StrategyTar:
		success, err = i.RunCmdGetSuccess("tar", "-xvf", tmpOut.Name(), "-C", *opts.Destination)
	case StrategyZip:
		success, err = i.RunCmdGetSuccess("unzip", tmpOut.Name(), "-d", *opts.Destination)
	default:
		io.Copy(out, tmpOut)
		success = true
		err = nil
	}

	if !success {
		return errors.Join(fmt.Errorf("Failed to extract the downloaded file"), err)
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
	_, err := i.RunCmdGetSuccess("git", "-C", i.GetInstallDir(), "fetch")
	if err != nil {
		return false, err
	}
	output, err := i.RunCmdGetOutput("git", "-C", i.GetInstallDir(), "status", "-uno")
	if err != nil {
		return false, err
	}
	if strings.Contains(string(output), "Your branch is behind") {
		return true, nil
	}
	return false, nil
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
	return i.Info
}

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
			opts.Repository = &destination
		}
		if filename, ok := (*info.Opts)["download_filename"].(string); ok {
			opts.DownloadFilename = &filename
		}
		if platformDownloadFilenames, ok := (*info.Opts)["platform_download_filenames"].(map[string]*string); ok {
			opts.PlatformDownloadFilenames = &platform.PlatformMap[string]{
				MacOS:   platformDownloadFilenames["macos"],
				Linux:   platformDownloadFilenames["linux"],
				Windows: platformDownloadFilenames["windows"],
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
	resp, err := http.Get(latestReleaseUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// parse json
	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	jsonMap := make(map[string]any)
	err = json.Unmarshal(contents, &jsonMap)
	if err != nil {
		return "", err
	}
	// get the release tag
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
	return filepath.Join(i.GetDestination(), filepath.Base(*i.Info.Name))
}

func NewGitHubReleaseInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *GitHubReleaseInstaller {
	i := &GitHubReleaseInstaller{
		Config: cfg,
		Info:   installer,
	}

	return i
}
