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
	return i.runOrStartContainer(false)
}

// Update implements IInstaller.
func (i *DockerInstaller) Update() error {
	image := *i.Info.Name
	containerName := i.GetContainerName()

	logger.Debug("Pulling updated image: %s", image)
	if err := i.RunCmdAsFile(fmt.Sprintf("docker pull %s", image)); err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	// Check if container exists before trying to remove
	exists := exec.Command("docker", "inspect", containerName).Run() == nil
	if exists {
		logger.Debug("Removing existing container: %s", containerName)
		_ = exec.Command("docker", "rm", "-f", containerName).Run()
	}

	logger.Debug("Running updated container: %s", containerName)
	return i.runOrStartContainer(true)
}

// CheckNeedsUpdate implements IInstaller.
func (i *DockerInstaller) CheckNeedsUpdate() (bool, error) {
	if i.HasCustomUpdateCheck() {
		return i.RunCustomUpdateCheck()
	}

	image := *i.Info.Name

	localDigest, err := i.getRepoDigestFromBeforePull(image)
	if err != nil {
		// If the image isn't present locally, we assume an update is needed
		logger.Debug("No local image found, assuming update needed")
		return true, nil
	}

	remoteDigest, err := i.getRemoteRepoDigest(image)
	if err != nil {
		return false, fmt.Errorf("failed to get remote image digest: %w", err)
	}

	logger.Debug("Local digest: %s", localDigest)
	logger.Debug("Remote digest: %s", remoteDigest)

	return localDigest != remoteDigest, nil
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
		if platformMap, ok := (*i.Info.Opts)["platform"].(map[string]*string); ok {
			opts.Platform = &platform.PlatformMap[string]{
				MacOS:   platformMap["macos"],
				Linux:   platformMap["linux"],
				Windows: platformMap["windows"],
			}
		}
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

// getRemoteRepoDigest fetches the remote repository digest for a Docker image.
func (i *DockerInstaller) getRemoteRepoDigest(image string) (string, error) {
	logger.Debug("Fetching remote image digest for: %s", image)
	cmd := exec.Command("docker", "manifest", "inspect", image)
	output, err := cmd.Output()
	if err != nil {
		logger.Debug("Docker manifest inspect failed: %v", err)
		return "", fmt.Errorf("docker manifest inspect failed: %w", err)
	}

	var osTarget, archTarget *string

	opts := i.GetOpts()
	if opts.Platform != nil {
		if resolved := opts.Platform.Resolve(); resolved != nil {
			parts := strings.SplitN(*resolved, "/", 2)
			if len(parts) == 2 {
				osTarget, archTarget = &parts[0], &parts[1]
			}
		}
	}

	if osTarget == nil {
		osTarget = platform.DockerOSMap.Resolve()
	}
	if archTarget == nil {
		archTarget = platform.DockerArchMap.Resolve()
	}
	if osTarget == nil || archTarget == nil {
		logger.Debug("Could not resolve platform for manifest digest: OS=%v, Arch=%v", osTarget, archTarget)
		return "", fmt.Errorf("could not resolve platform for manifest digest")
	}
	logger.Debug("Resolved OS: %s, Architecture: %s", *osTarget, *archTarget)

	digest, err := extractDigestFromManifest(output, *osTarget, *archTarget)
	if err == nil {
		return digest, nil
	}

	// fallback: architecture-only
	logger.Debug("Attempting fallback: match architecture only")
	digest, fallbackErr := extractDigestFromManifestAnyOS(output, *archTarget)
	if fallbackErr == nil {
		logger.Debug("Using fallback digest with architecture-only match: sha256:%s", digest)
		return digest, nil
	}

	return "", fmt.Errorf("no digest found for %s/%s or fallback: %w", *osTarget, *archTarget, err)
}

// extractDigestFromManifestAnyOS extracts the digest for a specific architecture from a Docker manifest list, ignoring the OS.
func extractDigestFromManifestAnyOS(jsonData []byte, archTarget string) (string, error) {
	var manifest DockerManifestList
	logger.Debug("Parsing manifest JSON data for architecture: %s", archTarget)
	if err := json.Unmarshal(jsonData, &manifest); err != nil {
		logger.Debug("Failed to parse manifest JSON: %v", err)
		return "", fmt.Errorf("failed to parse manifest JSON: %w", err)
	}
	for _, m := range manifest.Manifests {
		if m.Platform.Architecture == archTarget {
			logger.Debug("Found fallback digest for architecture: %s, Digest: %s", archTarget, m.Digest)
			return strings.TrimPrefix(m.Digest, "sha256:"), nil
		}
	}
	logger.Debug("No fallback digest found for architecture: %s", archTarget)
	return "", fmt.Errorf("no fallback digest found for arch: %s", archTarget)
}

// GetPlatformArchWithFallback attempts to determine the best architecture for a Docker image,
// considering a preferred architecture and a list of fallbacks.
// It inspects the manifest of a sample image ("ghcr.io/open-webui/open-webui:main") to check for available architectures.
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

// getRepoDigestFromBeforePull fetches the local repository digest for a Docker image before pulling.
func (i *DockerInstaller) getRepoDigestFromBeforePull(image string) (string, error) {
	logger.Debug("Checking local image digest before pull: %s", image)
	out, err := exec.Command("docker", "image", "inspect", "--format", "{{index .RepoDigests 0}}", image).Output()
	if err != nil {
		logger.Debug("Failed to get local image digest: %v", err)
		return "", err
	}
	parts := strings.Split(strings.TrimSpace(string(out)), "@")
	logger.Debug("Local image digest output: %s", out)
	if len(parts) != 2 {
		logger.Debug("Unexpected digest format before pull: %s", out)
		return "", fmt.Errorf("unexpected digest format before pull: %s", out)
	}

	digest := parts[1]
	digest, _ = strings.CutPrefix(digest, "sha256:")
	logger.Debug("Extracted local digest: %s", digest)
	return digest, nil
}
