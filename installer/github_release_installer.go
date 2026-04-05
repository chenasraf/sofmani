package installer

import (
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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
	// Strategy is the installation strategy to use (none, tar, zip, gzip).
	Strategy *GitHubReleaseInstallStrategy
	// GithubToken is the GitHub personal access token for authenticated API requests.
	// Supports environment variable expansion (e.g., "$GITHUB_TOKEN" or "${GITHUB_TOKEN}").
	GithubToken *string
	// ArchiveBinName is the name of the binary file inside the archive (tar/zip).
	// Use this when the filename inside the archive differs from the desired output bin_name.
	// If not set, falls back to bin_name (or the installer name).
	ArchiveBinName *string
	// ExtractTo, when set, switches the installer to "tree mode": the full archive contents
	// are extracted to this directory, preserving sibling files (lib/, share/, etc.) that
	// many toolchains rely on at runtime. Requires strategy 'tar' or 'zip'. When tree mode
	// is active, Destination and ArchiveBinName are ignored.
	ExtractTo *string
	// StripComponents drops this many leading path components from each archive entry, the
	// same way `tar --strip-components=N` does. Useful because release tarballs typically
	// wrap their contents in a single versioned directory. Only meaningful with ExtractTo.
	StripComponents *int
	// BinLinks lists binaries to expose from inside ExtractTo. On unix, each entry becomes
	// a symlink at Target pointing to Source; on Windows, the file is copied instead (since
	// symlinks require elevated privileges). Only meaningful with ExtractTo.
	BinLinks []GitHubReleaseBinLink
	// ExtractCommand is a user-provided shell command that performs the extraction when
	// Strategy is "custom". The command is run through Go template substitution with these
	// extra variables available (in addition to the usual .OS, .Arch, .Tag, ...):
	//   {{ .DownloadFile }}   - absolute path to the downloaded asset
	//   {{ .ExtractDir }}     - temp directory where the command should place extracted files
	//   {{ .Destination }}    - final destination directory
	//   {{ .BinName }}        - expected binary name (matches GetBinName())
	//   {{ .ArchiveBinName }} - the filename sofmani will copy from ExtractDir to Destination
	// After the command finishes, sofmani copies ExtractDir/ArchiveBinName to
	// Destination/BinName, the same way the tar and zip strategies do.
	ExtractCommand *string
}

// GitHubReleaseBinLink describes a single binary exposed from a tree-mode install.
type GitHubReleaseBinLink struct {
	// Source is the path to the binary inside the extracted tree. If relative, it is
	// resolved against ExtractTo; absolute paths are also accepted.
	Source string
	// Target is the absolute path where the symlink (or copied file, on Windows) is placed.
	Target string
}

// GitHubReleaseInstallStrategy represents the installation strategy for a GitHub release.
type GitHubReleaseInstallStrategy string

// Constants for GitHub release installation strategies.
const (
	GitHubReleaseInstallStrategyNone   GitHubReleaseInstallStrategy = "none"   // GitHubReleaseInstallStrategyNone means no special handling, just download the file.
	GitHubReleaseInstallStrategyTar    GitHubReleaseInstallStrategy = "tar"    // GitHubReleaseInstallStrategyTar means extract a tar archive.
	GitHubReleaseInstallStrategyZip    GitHubReleaseInstallStrategy = "zip"    // GitHubReleaseInstallStrategyZip means extract a zip archive.
	GitHubReleaseInstallStrategyGzip   GitHubReleaseInstallStrategy = "gzip"   // GitHubReleaseInstallStrategyGzip means decompress a single gzip-compressed file (not a tar archive).
	GitHubReleaseInstallStrategyCustom GitHubReleaseInstallStrategy = "custom" // GitHubReleaseInstallStrategyCustom runs a user-provided shell command to extract the asset.
)

// Validate validates the installer configuration.
func (i *GitHubReleaseInstaller) Validate() []ValidationError {
	errors := i.BaseValidate()
	info := i.GetData()
	opts := i.GetOpts()
	if opts.Repository == nil || len(*opts.Repository) == 0 {
		errors = append(errors, ValidationError{FieldName: "repository", Message: validationIsRequired(), InstallerName: *info.Name})
	}
	// In tree mode (extract_to set), destination is not required — bin_links handle
	// surfacing binaries on $PATH instead.
	if opts.ExtractTo == nil {
		if opts.Destination == nil || len(*opts.Destination) == 0 {
			errors = append(errors, ValidationError{FieldName: "destination", Message: validationIsRequired(), InstallerName: *info.Name})
		}
	}
	if opts.DownloadFilename == nil || len(*opts.DownloadFilename.Resolve()) == 0 {
		errors = append(errors, ValidationError{FieldName: "download_filename", Message: validationIsRequired(), InstallerName: *info.Name})
	} else if (*opts.DownloadFilename).Resolve() == nil || len(*(*opts.DownloadFilename).Resolve()) == 0 {
		errors = append(errors, ValidationError{FieldName: fmt.Sprintf("download_filename.%s", platform.GetPlatform()), Message: validationIsRequired(), InstallerName: *info.Name})
	}
	if opts.Strategy != nil {
		switch *opts.Strategy {
		case GitHubReleaseInstallStrategyNone,
			GitHubReleaseInstallStrategyTar,
			GitHubReleaseInstallStrategyZip,
			GitHubReleaseInstallStrategyGzip,
			GitHubReleaseInstallStrategyCustom:
			// valid
		default:
			errors = append(errors, ValidationError{FieldName: "strategy", Message: validationInvalidFormat(), InstallerName: *info.Name})
		}
	}
	// extract_command only makes sense with strategy: custom, and strategy: custom requires it.
	strategyIsCustom := opts.Strategy != nil && *opts.Strategy == GitHubReleaseInstallStrategyCustom
	hasExtractCommand := opts.ExtractCommand != nil && *opts.ExtractCommand != ""
	if strategyIsCustom && !hasExtractCommand {
		errors = append(errors, ValidationError{FieldName: "extract_command", Message: validationIsRequired(), InstallerName: *info.Name})
	}
	if hasExtractCommand && !strategyIsCustom {
		errors = append(errors, ValidationError{FieldName: "extract_command", Message: "extract_command requires strategy: custom", InstallerName: *info.Name})
	}
	if opts.ExtractTo != nil {
		// Tree mode requires an archive strategy — a single downloaded file has no tree to
		// extract. We check explicitly rather than relying on the Install-time error so
		// misconfigurations surface during validation.
		strategy := GitHubReleaseInstallStrategyNone
		if opts.Strategy != nil {
			strategy = *opts.Strategy
		}
		if strategy != GitHubReleaseInstallStrategyTar && strategy != GitHubReleaseInstallStrategyZip {
			errors = append(errors, ValidationError{FieldName: "strategy", Message: "extract_to requires strategy 'tar' or 'zip'", InstallerName: *info.Name})
		}
		if opts.StripComponents != nil && *opts.StripComponents < 0 {
			errors = append(errors, ValidationError{FieldName: "strip_components", Message: validationInvalidFormat(), InstallerName: *info.Name})
		}
		for idx, link := range opts.BinLinks {
			if link.Source == "" {
				errors = append(errors, ValidationError{FieldName: fmt.Sprintf("bin_links[%d].source", idx), Message: validationIsRequired(), InstallerName: *info.Name})
			} else if !filepath.IsAbs(link.Source) {
				// Relative sources are joined onto extract_to at install time; reject any
				// that try to escape the extracted tree with leading "..".
				cleaned := filepath.Clean(link.Source)
				if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
					errors = append(errors, ValidationError{FieldName: fmt.Sprintf("bin_links[%d].source", idx), Message: validationInvalidFormat(), InstallerName: *info.Name})
				}
			}
			if link.Target == "" {
				errors = append(errors, ValidationError{FieldName: fmt.Sprintf("bin_links[%d].target", idx), Message: validationIsRequired(), InstallerName: *info.Name})
			}
		}
	}
	return errors
}

// Install implements IInstaller.
func (i *GitHubReleaseInstaller) Install() error {
	opts := i.GetOpts()
	if opts.ExtractTo != nil {
		return i.installTree()
	}
	data := i.GetData()
	name := *data.Name
	tmpDir, err := os.MkdirTemp("", "sofmani")
	if err != nil {
		return err
	}
	tmpFile := fmt.Sprintf("%s/%s.download", tmpDir, name)
	logger.Debug("Created temp directory: %s", tmpDir)
	tmpOut, err := os.Create(tmpFile)
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
	var machineAliases map[string]string
	if i.Config.MachineAliases != nil {
		machineAliases = *i.Config.MachineAliases
	}
	templateVars := NewTemplateVars(tag, machineAliases)
	filename, err = ApplyTemplate(filename, templateVars, name)
	if err != nil {
		return fmt.Errorf("failed to apply template to filename: %w", err)
	}
	downloadUrl := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", *opts.Repository, tag, filename)
	logger.Debug("Downloading file: %s", filename)
	logger.Debug("Download URL: %s", downloadUrl)
	logger.Debug("Temp file: %s", tmpFile)

	req, err := http.NewRequest("GET", downloadUrl, nil)
	if err != nil {
		return err
	}
	if opts.GithubToken != nil && *opts.GithubToken != "" {
		logger.Debug("Using GitHub token for authentication")
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to download release asset: %s returned status %d", downloadUrl, resp.StatusCode)
	}

	n, err := io.Copy(tmpOut, resp.Body)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("no data was written to the file")
	}
	logger.Debug("Downloaded %d bytes to temp file", n)

	strategy := GitHubReleaseInstallStrategyNone

	if opts.Strategy != nil {
		strategy = *opts.Strategy
	}

	logger.Debug("Using strategy: %s", strategy)

	success := false

	outPath := filepath.Join(*opts.Destination, i.GetBinName())
	logger.Debug("Final destination: %s", outPath)

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
		logger.Debug("Strategy 'tar': extracting archive to %s", tmpDir)
		success, err = i.RunCmdGetSuccess("tar", "-xvf", tmpOut.Name(), "-C", tmpDir)
		if !success {
			return wrapExtractError("tar", tmpOut.Name(), err)
		}
		if err != nil {
			return err
		}
		logger.Debug("Strategy 'tar': copying binary '%s' to destination", i.GetArchiveBinName())
		success, err = i.CopyExtractedFile(out, tmpDir)
		if !success {
			return fmt.Errorf("failed to copy extracted file: %w", err)
		}
		if err != nil {
			return err
		}
	case GitHubReleaseInstallStrategyZip:
		logger.Debug("Strategy 'zip': extracting archive to %s", tmpDir)
		success, err = i.RunCmdGetSuccess("unzip", tmpOut.Name(), "-d", tmpDir)
		if !success {
			return wrapExtractError("zip", tmpOut.Name(), err)
		}
		if err != nil {
			return err
		}
		logger.Debug("Strategy 'zip': copying binary '%s' to destination", i.GetArchiveBinName())
		success, err = i.CopyExtractedFile(out, tmpDir)
		if !success {
			return fmt.Errorf("failed to copy extracted file: %w", err)
		}
		if err != nil {
			return err
		}
	case GitHubReleaseInstallStrategyGzip:
		logger.Debug("Strategy 'gzip': decompressing downloaded file to %s", outPath)
		if _, err = tmpOut.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek temp file: %w", err)
		}
		if err = decompressGzip(tmpOut, out); err != nil {
			return fmt.Errorf("failed to decompress gzip file: %w", err)
		}
		success = true
		err = nil
	case GitHubReleaseInstallStrategyCustom:
		logger.Debug("Strategy 'custom': running user extract_command against %s", tmpOut.Name())
		if opts.ExtractCommand == nil || *opts.ExtractCommand == "" {
			return fmt.Errorf("strategy 'custom' requires opts.extract_command")
		}
		extractVars := *templateVars
		extractVars.DownloadFile = tmpOut.Name()
		extractVars.ExtractDir = tmpDir
		extractVars.Destination = *opts.Destination
		extractVars.BinName = i.GetBinName()
		extractVars.ArchiveBinName = i.GetArchiveBinName()
		if err = i.runCustomExtract(*opts.ExtractCommand, &extractVars); err != nil {
			return fmt.Errorf("custom extract failed: %w", err)
		}
		logger.Debug("Strategy 'custom': copying binary '%s' to destination", i.GetArchiveBinName())
		success, err = i.CopyExtractedFile(out, tmpDir)
		if !success {
			return fmt.Errorf("failed to copy extracted file: %w", err)
		}
		if err != nil {
			return err
		}
	default:
		logger.Debug("Strategy 'none': copying downloaded file directly to destination")
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
	logger.Debug("Set executable permissions on %s", outPath)

	err = i.UpdateCache(tag)
	if err != nil {
		return err
	}

	logger.Debug("Installation complete: %s -> %s", filename, outPath)
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
	opts := i.GetOpts()
	if opts.ExtractTo != nil {
		// Tree mode: the install is present iff the extracted tree exists AND every
		// declared bin_link target exists. Removing a symlink from ~/.local/bin should
		// trigger reinstall so the user's expected entry points come back.
		logger.Debug("Checking if %s is installed at %s (tree mode)", *i.Info.Name, *opts.ExtractTo)
		exists, err := utils.PathExists(*opts.ExtractTo)
		if err != nil || !exists {
			return false, err
		}
		for _, link := range opts.BinLinks {
			exists, err := utils.PathExists(link.Target)
			if err != nil || !exists {
				return false, err
			}
		}
		return true, nil
	}
	logger.Debug("Checking if %s is installed on %s", *i.Info.Name, filepath.Join(i.GetInstallDir(), i.GetBinName()))
	return utils.PathExists(filepath.Join(i.GetInstallDir(), i.GetBinName()))
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

// GetArchiveBinName returns the name of the binary file inside the archive.
// It uses ArchiveBinName from opts if provided, otherwise falls back to GetBinName().
func (i *GitHubReleaseInstaller) GetArchiveBinName() string {
	opts := i.GetOpts()
	if opts.ArchiveBinName != nil {
		return *opts.ArchiveBinName
	}
	return i.GetBinName()
}

// runCustomExtract runs a user-provided extract command through the platform's
// default shell. The command is first rendered with ApplyTemplate so users can
// reference {{ .DownloadFile }}, {{ .ExtractDir }}, {{ .Destination }},
// {{ .BinName }}, {{ .ArchiveBinName }}, and all the usual template variables
// (.OS, .Arch, .Tag, ...).
func (i *GitHubReleaseInstaller) runCustomExtract(command string, vars *TemplateVars) error {
	rendered, err := ApplyTemplate(command, vars, *i.Info.Name)
	if err != nil {
		return fmt.Errorf("failed to render extract_command template: %w", err)
	}
	logger.Debug("Custom extract command: %s", rendered)
	shell := utils.GetOSShell(i.GetData().EnvShell)
	args := utils.GetOSShellArgs(rendered)
	success, err := i.RunCmdGetSuccessPassThrough(shell, args...)
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("extract_command exited non-zero")
	}
	return nil
}

// decompressGzip reads a gzip-compressed stream from src and writes the
// decompressed bytes to dst. It is used by the "gzip" github-release strategy
// for single-file gzipped assets (i.e. not tarballs).
func decompressGzip(src io.Reader, dst io.Writer) error {
	gr, err := gzip.NewReader(src)
	if err != nil {
		return fmt.Errorf("not a valid gzip stream: %w", err)
	}
	defer func() {
		if cerr := gr.Close(); cerr != nil {
			logger.Warn("failed to close gzip reader: %v", cerr)
		}
	}()
	n, err := io.Copy(dst, gr)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("no data was written to the output file")
	}
	return nil
}

// wrapExtractError produces a helpful error for a failed tar/zip extraction.
// If the underlying archive tool exited non-zero but returned no Go error, we
// sniff the file's magic bytes to detect the "single gzipped binary shipped
// as .gz" case (common on GitHub releases) and tell the user to try the
// "gzip" strategy instead.
func wrapExtractError(kind string, path string, cause error) error {
	hint := ""
	if kind == "tar" && isGzipFile(path) && !isTarGzFile(path) {
		hint = " (file looks like a plain gzip-compressed binary, not a tarball — try strategy: gzip)"
	}
	if cause == nil {
		return fmt.Errorf("failed to extract %s file: archive tool exited non-zero%s", kind, hint)
	}
	return fmt.Errorf("failed to extract %s file%s: %w", kind, hint, cause)
}

// isGzipFile returns true if the file at path starts with the gzip magic
// bytes (0x1f, 0x8b).
func isGzipFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	var header [2]byte
	n, err := io.ReadFull(f, header[:])
	if err != nil || n < 2 {
		return false
	}
	return header[0] == 0x1f && header[1] == 0x8b
}

// isTarGzFile returns true if the file at path is a gzip stream whose
// decompressed content begins with a tar header (checked via the "ustar"
// magic at offset 257). A plain gzipped binary will fail this check.
func isTarGzFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	gr, err := gzip.NewReader(f)
	if err != nil {
		return false
	}
	defer func() { _ = gr.Close() }()
	buf := make([]byte, 512)
	n, err := io.ReadFull(gr, buf)
	if err != nil && n < 512 {
		return false
	}
	// "ustar" magic lives at offset 257 in a tar header block.
	return string(buf[257:262]) == "ustar"
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
	tmpBinFile, err := os.Open(filepath.Join(tmpDir, i.GetArchiveBinName()))
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
			// Accept "gz" as a friendly alias for "gzip".
			if strat == "gz" {
				strat = GitHubReleaseInstallStrategyGzip
			}
			opts.Strategy = &strat
		}
		if token, ok := (*info.Opts)["github_token"].(string); ok {
			token = utils.GetRealPath(i.GetData().Environ(), token)
			opts.GithubToken = &token
		}
		if archiveBinName, ok := (*info.Opts)["archive_bin_name"].(string); ok {
			opts.ArchiveBinName = &archiveBinName
		}
		if extractCommand, ok := (*info.Opts)["extract_command"].(string); ok {
			opts.ExtractCommand = &extractCommand
		}
		if extractTo, ok := (*info.Opts)["extract_to"].(string); ok {
			extractTo = utils.GetRealPath(i.GetData().Environ(), extractTo)
			opts.ExtractTo = &extractTo
		}
		if raw, ok := (*info.Opts)["strip_components"]; ok {
			switch v := raw.(type) {
			case int:
				opts.StripComponents = &v
			case int64:
				n := int(v)
				opts.StripComponents = &n
			case float64:
				n := int(v)
				opts.StripComponents = &n
			}
		}
		if raw, ok := (*info.Opts)["bin_links"]; ok {
			if list, ok := raw.([]any); ok {
				for _, entry := range list {
					link, ok := parseBinLinkEntry(entry, i.GetData().Environ())
					if ok {
						opts.BinLinks = append(opts.BinLinks, link)
					}
				}
			}
		}
	}
	return opts
}

// parseBinLinkEntry converts a single YAML bin_links entry (a map) into a GitHubReleaseBinLink.
// It accepts both map[string]any (yaml.v3) and map[any]any (yaml.v2) shapes defensively.
func parseBinLinkEntry(entry any, env []string) (GitHubReleaseBinLink, bool) {
	link := GitHubReleaseBinLink{}
	get := func(key string) (string, bool) {
		switch m := entry.(type) {
		case map[string]any:
			if v, ok := m[key].(string); ok {
				return v, true
			}
		case map[any]any:
			if v, ok := m[key].(string); ok {
				return v, true
			}
		}
		return "", false
	}
	if s, ok := get("source"); ok {
		// Only expand env / ~ if absolute; relative sources are resolved against ExtractTo later.
		if filepath.IsAbs(s) || strings.HasPrefix(s, "~") || strings.Contains(s, "$") {
			s = utils.GetRealPath(env, s)
		}
		link.Source = s
	}
	if t, ok := get("target"); ok {
		link.Target = utils.GetRealPath(env, t)
	}
	if link.Source == "" && link.Target == "" {
		return link, false
	}
	return link, true
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
// In tree mode it returns extract_to; otherwise it falls back to destination.
func (i *GitHubReleaseInstaller) GetInstallDir() string {
	if opts := i.GetOpts(); opts.ExtractTo != nil {
		return *opts.ExtractTo
	}
	return i.GetDestination()
}

// installTree handles "tree mode" installs where the full archive contents are extracted
// into opts.ExtractTo and individual binaries are exposed via opts.BinLinks. The extracted
// tree is swapped into place atomically so an interrupted or failed install cannot leave
// a half-written directory behind, and a successful update fully replaces the previous
// version (no stale files from an old release linger).
func (i *GitHubReleaseInstaller) installTree() error {
	opts := i.GetOpts()
	data := i.GetData()
	name := *data.Name

	strategy := GitHubReleaseInstallStrategyNone
	if opts.Strategy != nil {
		strategy = *opts.Strategy
	}
	if strategy != GitHubReleaseInstallStrategyTar && strategy != GitHubReleaseInstallStrategyZip {
		return fmt.Errorf("extract_to requires strategy 'tar' or 'zip', got %q", strategy)
	}

	tmpDir, err := os.MkdirTemp("", "sofmani")
	if err != nil {
		return err
	}
	defer func() {
		if rerr := os.RemoveAll(tmpDir); rerr != nil {
			logger.Warn("failed to remove temp dir %s: %v", tmpDir, rerr)
		}
	}()

	tmpFile, tag, err := i.downloadRelease(tmpDir, name)
	if err != nil {
		return err
	}

	extractTo := *opts.ExtractTo
	stripComponents := 0
	if opts.StripComponents != nil {
		stripComponents = *opts.StripComponents
	}

	// Extract into a staging sibling so the old tree stays intact until we're ready to swap.
	staging := extractTo + ".sofmani-new"
	if err := os.RemoveAll(staging); err != nil {
		return fmt.Errorf("failed to clean staging dir %s: %w", staging, err)
	}
	if err := os.MkdirAll(staging, 0755); err != nil {
		return fmt.Errorf("failed to create staging dir %s: %w", staging, err)
	}

	switch strategy {
	case GitHubReleaseInstallStrategyTar:
		args := []string{"-xf", tmpFile, "-C", staging}
		if stripComponents > 0 {
			args = append(args, fmt.Sprintf("--strip-components=%d", stripComponents))
		}
		logger.Debug("Extracting tar to staging: tar %v", args)
		success, runErr := i.RunCmdGetSuccess("tar", args...)
		if runErr != nil || !success {
			_ = os.RemoveAll(staging)
			if runErr == nil {
				runErr = fmt.Errorf("tar exited with non-zero status")
			}
			return fmt.Errorf("failed to extract tar file: %w", runErr)
		}
	case GitHubReleaseInstallStrategyZip:
		logger.Debug("Extracting zip to staging: %s (strip=%d)", staging, stripComponents)
		if err := extractZipWithStrip(tmpFile, staging, stripComponents); err != nil {
			_ = os.RemoveAll(staging)
			return fmt.Errorf("failed to extract zip file: %w", err)
		}
	}

	// Atomically replace the old tree with the new one. We move the old tree aside first
	// so we can roll back if the rename fails halfway.
	backup := ""
	if _, err := os.Stat(extractTo); err == nil {
		backup = extractTo + ".sofmani-old"
		if err := os.RemoveAll(backup); err != nil {
			_ = os.RemoveAll(staging)
			return fmt.Errorf("failed to clean backup dir %s: %w", backup, err)
		}
		if err := os.Rename(extractTo, backup); err != nil {
			_ = os.RemoveAll(staging)
			return fmt.Errorf("failed to move existing tree aside: %w", err)
		}
	} else if !os.IsNotExist(err) {
		_ = os.RemoveAll(staging)
		return fmt.Errorf("failed to stat extract_to %s: %w", extractTo, err)
	} else {
		if err := os.MkdirAll(filepath.Dir(extractTo), 0755); err != nil {
			_ = os.RemoveAll(staging)
			return fmt.Errorf("failed to create parent of extract_to: %w", err)
		}
	}

	if err := os.Rename(staging, extractTo); err != nil {
		if backup != "" {
			// Roll back to the previous tree so the user isn't left with nothing.
			if rerr := os.Rename(backup, extractTo); rerr != nil {
				logger.Warn("failed to restore previous tree from %s: %v", backup, rerr)
			}
		}
		_ = os.RemoveAll(staging)
		return fmt.Errorf("failed to move staged tree into place: %w", err)
	}
	if backup != "" {
		if rerr := os.RemoveAll(backup); rerr != nil {
			logger.Warn("failed to remove old tree backup %s: %v", backup, rerr)
		}
	}
	logger.Debug("Extracted tree to %s", extractTo)

	for _, link := range opts.BinLinks {
		sourcePath := link.Source
		if !filepath.IsAbs(sourcePath) {
			sourcePath = filepath.Join(extractTo, sourcePath)
		}
		if err := installBinLink(sourcePath, link.Target); err != nil {
			return fmt.Errorf("failed to install bin link %s -> %s: %w", sourcePath, link.Target, err)
		}
		logger.Debug("Installed bin link %s -> %s", sourcePath, link.Target)
	}

	if err := i.UpdateCache(tag); err != nil {
		return err
	}
	logger.Debug("Tree install complete: %s", extractTo)
	return nil
}

// downloadRelease downloads the configured release asset to tmpDir and returns the on-disk
// path plus the resolved tag. It encapsulates the tag lookup, template application, HTTP
// fetch, and file write so both single-file and tree-mode installs can share it.
func (i *GitHubReleaseInstaller) downloadRelease(tmpDir, name string) (string, string, error) {
	opts := i.GetOpts()

	tag, err := i.GetLatestTag()
	if err != nil {
		return "", "", err
	}

	filename := i.GetFilename()
	if filename == "" {
		return "", "", fmt.Errorf("no download filename provided")
	}
	var machineAliases map[string]string
	if i.Config != nil && i.Config.MachineAliases != nil {
		machineAliases = *i.Config.MachineAliases
	}
	templateVars := NewTemplateVars(tag, machineAliases)
	filename, err = ApplyTemplate(filename, templateVars, name)
	if err != nil {
		return "", "", fmt.Errorf("failed to apply template to filename: %w", err)
	}

	tmpFile := filepath.Join(tmpDir, name+".download")
	out, err := os.Create(tmpFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		if cerr := out.Close(); cerr != nil {
			logger.Warn("failed to close tmpOut file: %v", cerr)
		}
	}()

	downloadUrl := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", *opts.Repository, tag, filename)
	logger.Debug("Downloading file: %s", filename)
	logger.Debug("Download URL: %s", downloadUrl)
	logger.Debug("Temp file: %s", tmpFile)

	req, err := http.NewRequest("GET", downloadUrl, nil)
	if err != nil {
		return "", "", err
	}
	if opts.GithubToken != nil && *opts.GithubToken != "" {
		logger.Debug("Using GitHub token for authentication")
		req.Header.Set("Authorization", "Bearer "+*opts.GithubToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logger.Warn("failed to close response body: %v", cerr)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("failed to download release asset: %s returned status %d", downloadUrl, resp.StatusCode)
	}

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return "", "", err
	}
	if n == 0 {
		return "", "", fmt.Errorf("no data was written to the file")
	}
	logger.Debug("Downloaded %d bytes to temp file", n)
	return tmpFile, tag, nil
}

// extractZipWithStrip extracts a zip archive into dest, dropping the first `strip` leading
// path components from each entry (mirroring `tar --strip-components=N`). We implement this
// in-process via archive/zip rather than shelling out to `unzip` because unzip has no
// native equivalent of strip-components and because archive/zip works on every platform
// sofmani supports (including Windows, where `unzip` is not always available).
func extractZipWithStrip(zipPath, dest string, strip int) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := r.Close(); cerr != nil {
			logger.Warn("failed to close zip reader: %v", cerr)
		}
	}()

	destClean := filepath.Clean(dest)
	for _, f := range r.File {
		parts := strings.Split(filepath.ToSlash(f.Name), "/")
		// Drop trailing empty segment from "dir/" style entries so strip counts real dirs.
		if len(parts) > 0 && parts[len(parts)-1] == "" {
			parts = parts[:len(parts)-1]
		}
		if len(parts) == 0 {
			continue
		}
		if len(parts) <= strip {
			// This entry lives entirely inside the stripped prefix — skip it.
			continue
		}
		rel := filepath.Join(parts[strip:]...)
		target := filepath.Join(destClean, rel)

		// Defend against zip-slip: the resolved target must stay inside dest.
		if target != destClean && !strings.HasPrefix(target, destClean+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		if err := writeZipFile(f, target); err != nil {
			return err
		}
	}
	return nil
}

// writeZipFile extracts a single zip entry to target, preserving its mode bits so that
// executable bits on unix-style archives survive the round-trip.
func writeZipFile(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() {
		if cerr := rc.Close(); cerr != nil {
			logger.Warn("failed to close zip entry %s: %v", f.Name, cerr)
		}
	}()
	mode := f.Mode().Perm()
	if mode == 0 {
		mode = 0644
	}
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, rc); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// installBinLink exposes a single binary from inside the extracted tree at `target`. On
// unix this is a symlink (so the binary keeps resolving its siblings via its real
// location); on Windows we fall back to copying the file because creating symlinks
// requires elevated privileges or developer mode.
func installBinLink(source, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	// Remove whatever is currently at target (file, broken symlink, or old symlink) so we
	// can replace it cleanly. Use Lstat so we don't follow the symlink.
	if _, err := os.Lstat(target); err == nil {
		if err := os.Remove(target); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if runtime.GOOS == "windows" {
		return copyFile(source, target)
	}
	return os.Symlink(source, target)
}

// copyFile is the Windows fallback for installBinLink.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
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
