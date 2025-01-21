package installer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/utils"
)

type GitHubReleaseInstaller struct {
	Config *appconfig.AppConfig
	Info   *appconfig.InstallerData
}

type GitHubReleaseOpts struct {
	Repository  *string
	Destination *string
	Filename    *string
}

// Install implements IInstaller.
func (i *GitHubReleaseInstaller) Install() error {
	opts := i.GetOpts()

	out, err := os.Create(fmt.Sprintf("%s/%s", *opts.Destination, *i.GetData().Name))
	defer out.Close()
	if err != nil {
		return err
	}

	tag, err := i.GetLatestTag()
	if err != nil {
		return err
	}

	filename := strings.ReplaceAll(*opts.Filename, "{tag}", tag)
	downloadUrl := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", *opts.Repository, tag, filename)
	resp, err := http.Get(downloadUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("No data was written to the file")
	}

	return nil
}

// Update implements IInstaller.
func (i *GitHubReleaseInstaller) Update() error {
	return i.Install()
}

// CheckNeedsUpdate implements IInstaller.
func (i *GitHubReleaseInstaller) CheckNeedsUpdate() (error, bool) {
	if i.GetData().CheckHasUpdate != nil {
		return utils.RunCmdGetSuccess(i.Info.Environ(), utils.GetOSShell(i.GetData().EnvShell), utils.GetOSShellArgs(*i.GetData().CheckHasUpdate)...)
	}
	err, _ := utils.RunCmdGetSuccess(i.Info.Environ(), "git", "-C", i.GetInstallDir(), "fetch")
	if err != nil {
		return err, false
	}
	output, err := utils.RunCmdGetOutput(i.Info.Environ(), "git", "-C", i.GetInstallDir(), "status", "-uno")
	if err != nil {
		return err, false
	}
	if strings.Contains(string(output), "Your branch is behind") {
		return nil, true
	}
	return nil, false
}

// CheckIsInstalled implements IInstaller.
func (i *GitHubReleaseInstaller) CheckIsInstalled() (error, bool) {
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
		if filename, ok := (*info.Opts)["filename"].(string); ok {
			opts.Filename = &filename
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
