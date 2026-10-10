package installer

import (
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	// resolvedToken caches the GitHub token so github_token_command runs at most once per run.
	resolvedToken *string
}

// GitHubReleaseOpts represents options for the GitHubReleaseInstaller.
type GitHubReleaseOpts struct {
	// Repository is the GitHub repository (e.g., "owner/repo").
	Repository *string
	// Version pins the installer to a release tag, exactly as it is named on GitHub
	// (e.g. "v1.2.3"). Without it, the repository's latest release is installed.
	Version *string
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
	// GithubTokenCommand is a shell command whose standard output is used as the GitHub token,
	// for tokens kept in a secret store rather than in the manifest or environment. It runs the
	// first time a request needs the token, including the latest-release lookup of an update check,
	// and only when GithubToken is empty.
	GithubTokenCommand *string
	// PreDownload and PostDownload are shell hooks run before and after the release asset is
	// downloaded. PreExtract and PostExtract run before and after it is extracted, and are skipped
	// with strategy "none". All four support the same template variables as ExtractCommand.
	PreDownload  *string
	PostDownload *string
	PreExtract   *string
	PostExtract  *string
	// ArchiveBinName is the name of the binary file inside the archive (tar/zip).
	// Use this when the filename inside the archive differs from the desired output bin_name.
	// Accepts either a string or a per-platform map. Supports Go template syntax with the
	// usual variables ({{ .Tag }}, {{ .Version }}, {{ .Arch }}, {{ .OS }}, ...).
	// If not set, falls back to bin_name (or the installer name).
	ArchiveBinName *platform.PlatformMap[string]
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
	if opts.DownloadFilename == nil {
		errors = append(errors, ValidationError{FieldName: "download_filename", Message: validationIsRequired(), InstallerName: *info.Name})
	} else if info.Platforms.GetShouldRunOnOS(platform.GetPlatform()) && (opts.DownloadFilename.Resolve() == nil || len(*opts.DownloadFilename.Resolve()) == 0) {
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
	tmpDir, err := os.MkdirTemp("", "sofmani")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	logger.Debug("Created temp directory: %s", tmpDir)

	err = os.MkdirAll(*opts.Destination, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", *opts.Destination, err)
	}

	vars, err := i.downloadRelease(tmpDir, *opts.Destination)
	if err != nil {
		return err
	}
	downloadFile := vars.DownloadFile

	strategy := GitHubReleaseInstallStrategyNone
	if opts.Strategy != nil {
		strategy = *opts.Strategy
	}
	logger.Debug("Using strategy: %s", strategy)

	outPath := filepath.Join(*opts.Destination, vars.BinName)
	logger.Debug("Final destination: %s", outPath)

	// Remove existing file first to avoid "text file busy" error on Linux
	// when updating a running executable
	if err := os.Remove(outPath); err != nil && !os.IsNotExist(err) {
		logger.Debug("Could not remove existing file: %v", err)
	}

	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outPath, err)
	}
	defer func() {
		if cerr := out.Close(); cerr != nil {
			logger.Warn("failed to close output file: %v", cerr)
		}
	}()

	extracts := strategy != GitHubReleaseInstallStrategyNone
	if extracts {
		if err := i.runHook("pre_extract", opts.PreExtract, vars); err != nil {
			return err
		}
	}

	switch strategy {
	case GitHubReleaseInstallStrategyTar:
		logger.Debug("Strategy 'tar': extracting archive to %s", tmpDir)
		success, err := i.RunCmdGetSuccess("tar", "-xvf", downloadFile, "-C", tmpDir)
		if !success {
			return wrapExtractError("tar", downloadFile, err)
		}
		if err != nil {
			return fmt.Errorf("failed to extract tar archive %s: %w", downloadFile, err)
		}
	case GitHubReleaseInstallStrategyZip:
		logger.Debug("Strategy 'zip': extracting archive to %s", tmpDir)
		success, err := i.RunCmdGetSuccess("unzip", downloadFile, "-d", tmpDir)
		if !success {
			return wrapExtractError("zip", downloadFile, err)
		}
		if err != nil {
			return fmt.Errorf("failed to extract zip archive %s: %w", downloadFile, err)
		}
	case GitHubReleaseInstallStrategyGzip:
		logger.Debug("Strategy 'gzip': decompressing downloaded file to %s", outPath)
		if err := copyDownload(downloadFile, out, decompressGzip); err != nil {
			return fmt.Errorf("failed to decompress gzip file %s to %s: %w", downloadFile, outPath, err)
		}
	case GitHubReleaseInstallStrategyCustom:
		logger.Debug("Strategy 'custom': running user extract_command against %s", downloadFile)
		if opts.ExtractCommand == nil || *opts.ExtractCommand == "" {
			return fmt.Errorf("strategy 'custom' requires opts.extract_command")
		}
		if err := i.runCustomExtract(*opts.ExtractCommand, vars); err != nil {
			return fmt.Errorf("custom extract failed (download=%s, extract_dir=%s): %w", downloadFile, tmpDir, err)
		}
	default:
		logger.Debug("Strategy 'none': copying downloaded file directly to destination")
		copyAll := func(src io.Reader, dst io.Writer) error {
			_, err := io.Copy(dst, src)
			return err
		}
		if err := copyDownload(downloadFile, out, copyAll); err != nil {
			return fmt.Errorf("failed to copy downloaded file %s to %s: %w", downloadFile, outPath, err)
		}
	}

	if extracts {
		if err := i.runHook("post_extract", opts.PostExtract, vars); err != nil {
			return err
		}
	}

	switch strategy {
	case GitHubReleaseInstallStrategyTar, GitHubReleaseInstallStrategyZip, GitHubReleaseInstallStrategyCustom:
		logger.Debug("Strategy '%s': copying binary '%s' to destination", strategy, vars.ArchiveBinName)
		success, err := i.CopyExtractedFile(out, tmpDir, vars)
		if !success {
			return fmt.Errorf("failed to copy extracted file %s from %s to %s: %w", vars.ArchiveBinName, tmpDir, outPath, err)
		}
		if err != nil {
			return fmt.Errorf("failed to copy extracted file %s to %s: %w", vars.ArchiveBinName, outPath, err)
		}
	}

	// Make the file executable
	if err = os.Chmod(outPath, 0755); err != nil {
		return fmt.Errorf("failed to make file %s executable: %w", outPath, err)
	}
	logger.Debug("Set executable permissions on %s", outPath)

	err = i.UpdateCache(vars.Tag)
	if err != nil {
		return err
	}

	logger.Debug("Installation complete: %s -> %s", downloadFile, outPath)
	return nil
}

// copyDownload opens the downloaded asset at path and hands it to write along with dst.
func copyDownload(path string, dst io.Writer, write func(src io.Reader, dst io.Writer) error) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	return write(in, dst)
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
	wanted, err := i.GetTag()
	if err != nil {
		return false, err
	}
	if wanted != cachedTag {
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
// It uses ArchiveBinName from opts if provided (resolved for the current platform when
// a per-platform map was supplied), otherwise falls back to GetBinName(). When vars is
// non-nil, the resolved value is rendered through ApplyTemplate so template tokens like
// {{ .Tag }} or {{ .Arch }} are substituted.
func (i *GitHubReleaseInstaller) GetArchiveBinName(vars *TemplateVars) string {
	opts := i.GetOpts()
	value := ""
	if opts.ArchiveBinName != nil {
		if resolved := opts.ArchiveBinName.Resolve(); resolved != nil {
			value = *resolved
		}
	}
	if value == "" {
		value = i.GetBinName()
	}
	if vars == nil {
		return value
	}
	rendered, err := ApplyTemplate(value, vars, *i.Info.Name)
	if err != nil {
		logger.Warn("failed to render archive_bin_name template: %v", err)
		return value
	}
	return rendered
}

// runCustomExtract runs a user-provided extract command through the platform's
// default shell. The command is first rendered with ApplyTemplate so users can
// reference {{ .DownloadFile }}, {{ .ExtractDir }}, {{ .Destination }},
// {{ .BinName }}, {{ .ArchiveBinName }}, and all the usual template variables
// (.OS, .Arch, .Tag, ...).
func (i *GitHubReleaseInstaller) runCustomExtract(command string, vars *TemplateVars) error {
	rendered, err := ApplyTemplate(command, vars, *i.Info.Name)
	if err != nil {
		return fmt.Errorf("failed to render extract_command template %q: %w", command, err)
	}
	logger.Debug("Custom extract command: %s", rendered)
	shell := utils.GetOSShell(i.GetData().EnvShell)
	args := utils.GetOSShellArgs(rendered)
	success, err := i.RunCmdGetSuccessPassThrough(shell, args...)
	if err != nil {
		return fmt.Errorf("extract_command failed (%q): %w", rendered, err)
	}
	if !success {
		return fmt.Errorf("extract_command exited non-zero: %q", rendered)
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
// vars is used to render template tokens inside archive_bin_name; pass nil to skip templating.
func (i *GitHubReleaseInstaller) CopyExtractedFile(out *os.File, tmpDir string, vars *TemplateVars) (bool, error) {
	binFile, err := os.Create(out.Name())
	if err != nil {
		return false, fmt.Errorf("failed to create output file %s: %w", out.Name(), err)
	}
	defer func() {
		if cerr := binFile.Close(); cerr != nil {
			logger.Warn("failed to close binFile %s: %v", binFile.Name(), cerr)
		}
	}()
	archiveBin := i.GetArchiveBinName(vars)
	srcPath := filepath.Join(tmpDir, archiveBin)
	tmpBinFile, err := os.Open(srcPath)
	if err != nil {
		return false, fmt.Errorf("failed to open extracted binary at %s (archive_bin_name=%q, extract_dir=%s): %w", srcPath, archiveBin, tmpDir, err)
	}
	logger.Debug("Copying file %s to %s", tmpBinFile.Name(), binFile.Name())

	n, err := io.Copy(binFile, tmpBinFile)
	if err != nil {
		return false, fmt.Errorf("failed to copy %s to %s: %w", srcPath, binFile.Name(), err)
	}
	if n == 0 {
		return false, fmt.Errorf("no data was written to %s (source %s is empty)", binFile.Name(), srcPath)
	}
	return true, nil
}

// GetCachedTag retrieves the cached tag for the release from the cache directory.
func (i *GitHubReleaseInstaller) GetCachedTag() (string, error) {
	logger.Debug("Getting cached tag for %s", *i.Info.Name)
	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve cache directory: %w", err)
	}
	cacheFile := fmt.Sprintf("%s/%s", cacheDir, *i.Info.Name)
	exists, err := utils.PathExists(cacheFile)
	if err != nil {
		return "", fmt.Errorf("failed to stat cache file %s: %w", cacheFile, err)
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
		return fmt.Errorf("failed to resolve cache directory: %w", err)
	}
	cacheFile := fmt.Sprintf("%s/%s", cacheDir, *i.Info.Name)
	logger.Debug("Updating cache file %s with %s", cacheFile, tag)
	err = os.WriteFile(cacheFile, []byte(tag), 0644)
	if err != nil {
		return fmt.Errorf("failed to write cache file %s: %w", cacheFile, err)
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
		if version, ok := (*info.Opts)["version"].(string); ok {
			opts.Version = &version
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
		if command, ok := (*info.Opts)["github_token_command"].(string); ok {
			opts.GithubTokenCommand = &command
		}
		for key, target := range map[string]**string{
			"pre_download":  &opts.PreDownload,
			"post_download": &opts.PostDownload,
			"pre_extract":   &opts.PreExtract,
			"post_extract":  &opts.PostExtract,
		} {
			if hook, ok := (*info.Opts)[key].(string); ok {
				*target = &hook
			}
		}
		if raw, ok := (*info.Opts)["archive_bin_name"]; ok {
			opts.ArchiveBinName = platform.NewPlatformMap[string](raw)
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

// GetPinnedVersion implements IVersionPinned. The pin is the release tag to install.
func (i *GitHubReleaseInstaller) GetPinnedVersion() string {
	if version := i.GetOpts().Version; version != nil {
		return *version
	}
	return ""
}

// GetTag returns the release tag to install: the pinned opts.version when set, otherwise the
// repository's latest release tag.
func (i *GitHubReleaseInstaller) GetTag() (string, error) {
	if tag := i.GetPinnedVersion(); tag != "" {
		logger.Debug("%s is pinned to release %s", *i.Info.Name, tag)
		return tag, nil
	}
	return i.GetLatestTag()
}

func (i *GitHubReleaseInstaller) GetLatestTag() (string, error) {
	opts := i.GetOpts()
	latestReleaseUrl := fmt.Sprintf("%s/repos/%s/releases/latest", githubAPIBaseURL, *opts.Repository)
	logger.Debug("Getting latest release from %s", latestReleaseUrl)

	req, err := http.NewRequest("GET", latestReleaseUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build request for %s: %w", latestReleaseUrl, err)
	}
	token, err := i.githubToken()
	if err != nil {
		return "", err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release from %s: %w", latestReleaseUrl, err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Warn("Failed to close response body: %v", err)
		}
	}()
	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response from %s: %w", latestReleaseUrl, err)
	}
	jsonMap := make(map[string]any)
	err = json.Unmarshal(contents, &jsonMap)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON response from %s: %w", latestReleaseUrl, err)
	}
	tag, ok := jsonMap["tag_name"].(string)
	if !ok || tag == "" {
		logger.Warn("Invalid GitHub API response from %s: %s", latestReleaseUrl, string(contents))
		if msg, ok := jsonMap["message"].(string); ok {
			return "", fmt.Errorf("GitHub API error for %s: %s", latestReleaseUrl, msg)
		}
		return "", fmt.Errorf("no releases found for repository %s (queried %s)", *opts.Repository, latestReleaseUrl)
	}
	logger.Debug("Latest release is %s", tag)
	return tag, nil
}

// GetFilename returns the filename to download from the release, resolved for the current platform.
func (i *GitHubReleaseInstaller) GetFilename() string {
	opts := i.GetOpts()
	if opts.DownloadFilename != nil {
		resolved := opts.DownloadFilename.Resolve()
		if resolved != nil {
			return *resolved
		}
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

	strategy := GitHubReleaseInstallStrategyNone
	if opts.Strategy != nil {
		strategy = *opts.Strategy
	}
	if strategy != GitHubReleaseInstallStrategyTar && strategy != GitHubReleaseInstallStrategyZip {
		return fmt.Errorf("extract_to requires strategy 'tar' or 'zip', got %q", strategy)
	}

	tmpDir, err := os.MkdirTemp("", "sofmani")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if rerr := os.RemoveAll(tmpDir); rerr != nil {
			logger.Warn("failed to remove temp dir %s: %v", tmpDir, rerr)
		}
	}()

	extractTo := *opts.ExtractTo
	vars, err := i.downloadRelease(tmpDir, extractTo)
	if err != nil {
		return err
	}
	tmpFile := vars.DownloadFile

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

	// The extract hooks see the staging tree, so post_extract can adjust it before it is swapped in.
	vars.ExtractDir = staging
	if err := i.runHook("pre_extract", opts.PreExtract, vars); err != nil {
		_ = os.RemoveAll(staging)
		return err
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
			return fmt.Errorf("failed to extract tar file %s to %s: %w", tmpFile, staging, runErr)
		}
	case GitHubReleaseInstallStrategyZip:
		logger.Debug("Extracting zip to staging: %s (strip=%d)", staging, stripComponents)
		if err := extractZipWithStrip(tmpFile, staging, stripComponents); err != nil {
			_ = os.RemoveAll(staging)
			return fmt.Errorf("failed to extract zip file %s to %s: %w", tmpFile, staging, err)
		}
	}

	if err := i.runHook("post_extract", opts.PostExtract, vars); err != nil {
		_ = os.RemoveAll(staging)
		return err
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
			return fmt.Errorf("failed to move existing tree %s aside to %s: %w", extractTo, backup, err)
		}
	} else if !os.IsNotExist(err) {
		_ = os.RemoveAll(staging)
		return fmt.Errorf("failed to stat extract_to %s: %w", extractTo, err)
	} else {
		if err := os.MkdirAll(filepath.Dir(extractTo), 0755); err != nil {
			_ = os.RemoveAll(staging)
			return fmt.Errorf("failed to create parent of extract_to %s: %w", extractTo, err)
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
		return fmt.Errorf("failed to move staged tree %s to %s: %w", staging, extractTo, err)
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

	if err := i.UpdateCache(vars.Tag); err != nil {
		return err
	}
	logger.Debug("Tree install complete: %s", extractTo)
	return nil
}

// downloadRelease downloads the configured release asset into tmpDir, running the
// pre_download and post_download hooks around it, and returns the template variables for the
// rest of the install: the resolved tag, DownloadFile, and ExtractDir (tmpDir), Destination,
// BinName and ArchiveBinName. Both single-file and tree-mode installs share it.
func (i *GitHubReleaseInstaller) downloadRelease(tmpDir, destination string) (*TemplateVars, error) {
	opts := i.GetOpts()
	name := *i.Info.Name
	tag, err := i.GetTag()
	if err != nil {
		return nil, err
	}

	filename := i.GetFilename()
	if filename == "" {
		return nil, fmt.Errorf("no download filename matched for the current platform (%s/%s)", runtime.GOOS, runtime.GOARCH)
	}
	var machineAliases map[string]string
	if i.Config != nil && i.Config.MachineAliases != nil {
		machineAliases = *i.Config.MachineAliases
	}
	vars := NewTemplateVars(tag, machineAliases)
	rawFilename := filename
	filename, err = ApplyTemplate(filename, vars, name)
	if err != nil {
		return nil, fmt.Errorf("failed to apply template to download_filename %q: %w", rawFilename, err)
	}

	tmpFile := filepath.Join(tmpDir, name+".download")
	vars.DownloadFile = tmpFile
	vars.ExtractDir = tmpDir
	vars.Destination = destination
	vars.BinName = i.GetBinName()
	vars.ArchiveBinName = i.GetArchiveBinName(vars)

	if err := i.runHook("pre_download", opts.PreDownload, vars); err != nil {
		return nil, err
	}

	out, err := os.Create(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file %s: %w", tmpFile, err)
	}
	logger.Debug("Temp file: %s", tmpFile)
	n, err := i.downloadAsset(out, tag, filename)
	// Closed here rather than deferred so post_download sees the complete file.
	if cerr := out.Close(); cerr != nil && err == nil {
		err = fmt.Errorf("failed to close temporary file %s: %w", tmpFile, cerr)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to download %s to %s: %w", filename, tmpFile, err)
	}
	logger.Debug("Downloaded %d bytes to temp file", n)

	if err := i.runHook("post_download", opts.PostDownload, vars); err != nil {
		return nil, err
	}
	return vars, nil
}

// runHook runs one of the opts shell hooks (pre_download, post_extract, ...) after rendering vars
// into it. An unset hook is a no-op.
func (i *GitHubReleaseInstaller) runHook(field string, command *string, vars *TemplateVars) error {
	if command == nil || *command == "" {
		return nil
	}
	rendered, err := ApplyTemplate(*command, vars, *i.Info.Name)
	if err != nil {
		return fmt.Errorf("failed to render %s template %q: %w", field, *command, err)
	}
	logger.Debug("Running %s hook for %s", field, *i.Info.Name)
	data := i.GetData()
	if err := i.RunCmdPassThrough(utils.GetOSShell(data.EnvShell), utils.GetOSShellArgs(rendered)...); err != nil {
		return fmt.Errorf("%s hook failed for %s: %w", field, *i.Info.Name, err)
	}
	return nil
}

// githubToken returns the token for GitHub API requests, or "" for anonymous requests. A
// non-empty github_token wins over github_token_command, so a token exported in the environment
// (e.g. in CI) skips the secret store. The command runs on the first call only; its trimmed
// output is reused afterwards.
func (i *GitHubReleaseInstaller) githubToken() (string, error) {
	if i.resolvedToken != nil {
		return *i.resolvedToken, nil
	}
	opts := i.GetOpts()
	token := ""
	switch {
	case opts.GithubToken != nil && *opts.GithubToken != "":
		token = *opts.GithubToken
	case opts.GithubTokenCommand != nil && *opts.GithubTokenCommand != "":
		logger.Debug("Running github_token_command for %s", *i.Info.Name)
		data := i.GetData()
		out, err := utils.RunCmdGetOutputPassThrough(data.Environ(), utils.GetOSShell(data.EnvShell), utils.GetOSShellArgs(i.applyTemplate(*opts.GithubTokenCommand))...)
		if err != nil {
			return "", fmt.Errorf("github_token_command failed for %s: %w", *i.Info.Name, err)
		}
		token = strings.TrimSpace(string(out))
		if token == "" {
			return "", fmt.Errorf("github_token_command for %s printed no token", *i.Info.Name)
		}
	}
	i.resolvedToken = &token
	return token, nil
}

// githubAPIBaseURL and githubWebBaseURL are variables so tests can point them at local servers.
var (
	githubAPIBaseURL = "https://api.github.com"
	githubWebBaseURL = "https://github.com"
)

// downloadAsset writes the release asset named filename from tag to out and returns the number
// of bytes written. With a token, it goes through the releases API: github.com download URLs
// ignore bearer tokens, so private repositories answer 404 there. Without a token, it uses the
// github.com download URL, which costs no API rate limit.
func (i *GitHubReleaseInstaller) downloadAsset(out io.Writer, tag, filename string) (int64, error) {
	opts := i.GetOpts()
	repo := *opts.Repository
	logger.Debug("Downloading file: %s", filename)

	token, err := i.githubToken()
	if err != nil {
		return 0, err
	}
	var req *http.Request
	if token != "" {
		logger.Debug("Using GitHub token for authentication")
		assetUrl, err := i.findReleaseAssetUrl(tag, filename, token)
		if err != nil {
			return 0, err
		}
		req, err = newGitHubAPIRequest(assetUrl, token)
		if err != nil {
			return 0, err
		}
		req.Header.Set("Accept", "application/octet-stream")
	} else {
		downloadUrl := fmt.Sprintf("%s/%s/releases/download/%s/%s", githubWebBaseURL, repo, tag, filename)
		var err error
		req, err = http.NewRequest("GET", downloadUrl, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to build request for %s: %w", downloadUrl, err)
		}
	}
	logger.Debug("Download URL: %s", req.URL)

	// The asset endpoint redirects to a signed URL on another host, which rejects requests that
	// carry an Authorization header. http.Client drops that header on cross-host redirects, so
	// no CheckRedirect is set here.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to download release asset from %s: %w", req.URL, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logger.Warn("failed to close response body: %v", cerr)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("failed to download release asset: %s returned status %d", req.URL, resp.StatusCode)
	}

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return n, fmt.Errorf("failed to write downloaded asset from %s: %w", req.URL, err)
	}
	if n == 0 {
		return 0, fmt.Errorf("no data was received from %s", req.URL)
	}
	return n, nil
}

// findReleaseAssetUrl looks up the release for tag and returns the API URL of its asset named
// filename.
func (i *GitHubReleaseInstaller) findReleaseAssetUrl(tag, filename, token string) (string, error) {
	opts := i.GetOpts()
	repo := *opts.Repository
	releaseUrl := fmt.Sprintf("%s/repos/%s/releases/tags/%s", githubAPIBaseURL, repo, url.PathEscape(tag))
	logger.Debug("Getting release from %s", releaseUrl)

	req, err := newGitHubAPIRequest(releaseUrl, token)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch release from %s: %w", releaseUrl, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logger.Warn("failed to close response body: %v", cerr)
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("failed to fetch release: %s returned status %d", releaseUrl, resp.StatusCode)
	}

	var release struct {
		Assets []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse JSON response from %s: %w", releaseUrl, err)
	}

	names := make([]string, 0, len(release.Assets))
	for _, asset := range release.Assets {
		if asset.Name == filename {
			return fmt.Sprintf("%s/repos/%s/releases/assets/%d", githubAPIBaseURL, repo, asset.ID), nil
		}
		names = append(names, asset.Name)
	}
	return "", fmt.Errorf("release %s of %s has no asset named %q (available: %s)", tag, repo, filename, strings.Join(names, ", "))
}

func newGitHubAPIRequest(requestUrl, token string) (*http.Request, error) {
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request for %s: %w", requestUrl, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return req, nil
}

// extractZipWithStrip extracts a zip archive into dest, dropping the first `strip` leading
// path components from each entry (mirroring `tar --strip-components=N`). We implement this
// in-process via archive/zip rather than shelling out to `unzip` because unzip has no
// native equivalent of strip-components and because archive/zip works on every platform
// sofmani supports (including Windows, where `unzip` is not always available).
func extractZipWithStrip(zipPath, dest string, strip int) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip %s: %w", zipPath, err)
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
			return fmt.Errorf("invalid file path in zip %s: entry %q would resolve outside %s", zipPath, f.Name, destClean)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", target, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", target, err)
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
		return fmt.Errorf("failed to open zip entry %s: %w", f.Name, err)
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
		return fmt.Errorf("failed to create %s: %w", target, err)
	}
	if _, err := io.Copy(out, rc); err != nil {
		_ = out.Close()
		return fmt.Errorf("failed to write zip entry %s to %s: %w", f.Name, target, err)
	}
	return out.Close()
}

// installBinLink exposes a single binary from inside the extracted tree at `target`. On
// unix this is a symlink (so the binary keeps resolving its siblings via its real
// location); on Windows we fall back to copying the file because creating symlinks
// requires elevated privileges or developer mode.
func installBinLink(source, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for %s: %w", target, err)
	}
	// Remove whatever is currently at target (file, broken symlink, or old symlink) so we
	// can replace it cleanly. Use Lstat so we don't follow the symlink.
	if _, err := os.Lstat(target); err == nil {
		if err := os.Remove(target); err != nil {
			return fmt.Errorf("failed to remove existing target %s: %w", target, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat target %s: %w", target, err)
	}
	if runtime.GOOS == "windows" {
		return copyFile(source, target)
	}
	if err := os.Symlink(source, target); err != nil {
		return fmt.Errorf("failed to create symlink %s -> %s: %w", target, source, err)
	}
	return nil
}

// copyFile is the Windows fallback for installBinLink.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source %s: %w", src, err)
	}
	defer func() { _ = in.Close() }()
	info, err := in.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source %s: %w", src, err)
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return fmt.Errorf("failed to open destination %s: %w", dst, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return fmt.Errorf("failed to copy %s to %s: %w", src, dst, err)
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
