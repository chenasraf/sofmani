package installer

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/platform"
)

// DockerInstaller is an installer for Docker images.
type DockerInstaller struct {
	InstallerBase
	// Config is the application configuration.
	Config *appconfig.AppConfig
	// Info is the installer data.
	Info *appconfig.InstallerData
}

// DockerOpts represents options for the DockerInstaller.
type DockerOpts struct {
	// Flags is a string of flags to pass to the `docker run` command.
	Flags *string
	// Platform is a platform-specific map of Docker platform strings (e.g., "linux/amd64").
	Platform *platform.PlatformMap[string]
	// SkipIfUnavailable indicates whether to skip installation if Docker is unavailable.
	SkipIfUnavailable *bool
}

// NewDockerInstaller creates a new DockerInstaller.
func NewDockerInstaller(cfg *appconfig.AppConfig, installer *appconfig.InstallerData) *DockerInstaller {
	return &DockerInstaller{
		InstallerBase: InstallerBase{Data: installer},
		Config:        cfg,
		Info:          installer,
	}
}

// Validate validates the installer configuration.
func (i *DockerInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	return errors
}

// Install implements IInstaller.
func (i *DockerInstaller) Install() error {
	if !isDockerAvailable() {
		if i.GetOpts().SkipIfUnavailable != nil && *i.GetOpts().SkipIfUnavailable {
			logger.Debug("Docker not available, skipping install")
			return nil
		}
		return fmt.Errorf("docker is not available")
	}
	return i.runOrStartContainer(false)
}

// Update implements IInstaller.
func (i *DockerInstaller) Update() error {
	if !isDockerAvailable() {
		if i.GetOpts().SkipIfUnavailable != nil && *i.GetOpts().SkipIfUnavailable {
			logger.Debug("Docker not available, skipping update")
			return nil
		}
		return fmt.Errorf("docker is not available")
	}

	image := *i.Info.Name
	containerName := i.GetContainerName()

	logger.Debug("Pulling updated image: %s", image)
	if err := i.RunCmdPassThrough("docker", "pull", image); err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	logger.Debug("Removing existing container: %s", containerName)
	err := i.RunCmdPassThrough("docker", "rm", "-f", containerName)
	if err != nil {
		logger.Debug("Failed to remove existing container: %s, error: %v", containerName, err)
		return fmt.Errorf("failed to remove existing container: %w", err)
	}

	logger.Debug("Running updated container: %s", containerName)
	return i.runOrStartContainer(true)
}

// CheckNeedsUpdate implements IInstaller.
func (i *DockerInstaller) CheckNeedsUpdate() (bool, error) {
	// Always assume an update is available
	return true, nil
}

// CheckIsInstalled implements IInstaller.
func (i *DockerInstaller) CheckIsInstalled() (bool, error) {
	if i.HasCustomInstallCheck() {
		return i.RunCustomInstallCheck()
	}

	containerName := i.GetContainerName()
	cmd := exec.Command("docker", "inspect", containerName)
	err := cmd.Run()
	return err == nil, nil
}

// GetData implements IInstaller.
func (i *DockerInstaller) GetData() *appconfig.InstallerData {
	return i.Info
}

// GetOpts returns the parsed options for the DockerInstaller.
func (i *DockerInstaller) GetOpts() *DockerOpts {
	opts := &DockerOpts{}
	if i.Info.Opts != nil {
		if flags, ok := (*i.Info.Opts)["flags"].(string); ok {
			opts.Flags = &flags
		}
		if raw, ok := (*i.Info.Opts)["platform"]; ok && raw != nil {
			opts.Platform = platform.NewPlatformMap[string](raw)
		}
	}
	if skip, ok := (*i.Info.Opts)["skip_if_unavailable"].(bool); ok {
		opts.SkipIfUnavailable = &skip
	}
	return opts
}

// GetContainerName returns the name of the Docker container.
// It uses the BinName from the installer data if provided, otherwise it uses the installer name.
func (i *DockerInstaller) GetContainerName() string {
	if i.Info.BinName != nil && len(*i.Info.BinName) > 0 {
		return *i.Info.BinName
	}
	return *i.Info.Name
}

// Helpers

// runOrStartContainer runs or starts a Docker container.
// If forceRun is true, it will always run a new container. Otherwise, it will start an existing container if found.
func (i *DockerInstaller) runOrStartContainer(forceRun bool) error {
	containerName := i.GetContainerName()
	image := *i.Info.Name
	opts := i.GetOpts()

	flags := "-d --restart always"
	if opts.Flags != nil {
		flat := strings.Join(strings.Fields(*opts.Flags), " ")
		flags += " " + flat
	}

	if !forceRun {
		exists := exec.Command("docker", "inspect", containerName).Run() == nil
		if exists {
			return i.RunCmdAsFile(fmt.Sprintf(`docker start "%s"`, containerName))
		}
	}

	return i.RunCmdAsFile(fmt.Sprintf(`docker run %s --name "%s" "%s"`, flags, containerName, image))
}

// DockerManifestList represents the structure of a Docker manifest list.
type DockerManifestList struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Manifests     []struct {
		Digest   string `json:"digest"`
		Platform struct {
			Architecture string `json:"architecture"`
			OS           string `json:"os"`
		} `json:"platform"`
	} `json:"manifests"`
}

// extractDigestFromManifest extracts the digest for a specific OS and architecture from a Docker manifest list.
func extractDigestFromManifest(jsonData []byte, osTarget, archTarget string) (string, error) {
	var manifest DockerManifestList
	logger.Debug("Parsing manifest JSON data for OS: %s, Arch: %s", osTarget, archTarget)
	if err := json.Unmarshal(jsonData, &manifest); err != nil {
		logger.Debug("Failed to parse manifest JSON: %v", err)
		return "", fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	for _, m := range manifest.Manifests {
		if m.Platform.OS == osTarget && m.Platform.Architecture == archTarget {
			return strings.TrimPrefix(m.Digest, "sha256:"), nil
		}
	}
	logger.Debug("No matching digest found for OS: %s, Arch: %s", osTarget, archTarget)
	logger.Debug("Available manifests: %v", manifest.Manifests)
	return "", fmt.Errorf("no digest found for %s/%s", osTarget, archTarget)
}

// GetPlatformArchWithFallback attempts to determine the best architecture for a Docker image,
// considering a preferred architecture and a list of fallbacks.
func GetPlatformArchWithFallback(preferred string, fallbacks ...string) string {
	image := "ghcr.io/open-webui/open-webui:main"
	cmd := exec.Command("docker", "manifest", "inspect", image)
	out, err := cmd.Output()
	if err != nil {
		return preferred
	}
	for _, arch := range append([]string{preferred}, fallbacks...) {
		if strings.Contains(string(out), fmt.Sprintf(`"architecture": "%s"`, arch)) {
			return arch
		}
	}
	return preferred
}

// isDockerAvailable checks if Docker is available on the system.
func isDockerAvailable() bool {
	err := exec.Command("docker", "info").Run()
	return err == nil
}
