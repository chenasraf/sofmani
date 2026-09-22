package installer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/utils"
)

// IVersionPinned is implemented by installers whose package manager can install an exact
// version. A pinned installer does not follow the newest release: sofmani records the
// version it installed, and reports an update only once the pin in the manifest changes.
type IVersionPinned interface {
	// GetPinnedVersion returns the pinned version, or an empty string when the installer
	// follows the newest available version.
	GetPinnedVersion() string
}

// cacheFileName builds a cache file name for an installer, escaping characters that are not
// safe for file names.
func cacheFileName(prefix string, name string) string {
	replacer := strings.NewReplacer(
		"/", "__",
		"\\", "__",
		":", "__",
		" ", "_",
	)
	return prefix + replacer.Replace(name)
}

// pinnedVersionCacheFile returns the path of the file holding the version last installed for
// the named installer.
func pinnedVersionCacheFile(name string) (string, error) {
	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, cacheFileName("version_", name)), nil
}

// ReadInstalledVersion returns the version sofmani last installed for the named installer,
// or an empty string when no version was recorded.
func ReadInstalledVersion(name string) string {
	file, err := pinnedVersionCacheFile(name)
	if err != nil {
		return ""
	}
	contents, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(contents))
}

// RecordInstalledVersion stores the version installed for the named installer.
func RecordInstalledVersion(name string, version string) error {
	file, err := pinnedVersionCacheFile(name)
	if err != nil {
		return err
	}
	return os.WriteFile(file, []byte(version), 0644)
}

// ClearInstalledVersion drops the version record for the named installer, which no longer
// describes what is installed once the installer stops being pinned.
func ClearInstalledVersion(name string) {
	file, err := pinnedVersionCacheFile(name)
	if err != nil {
		return
	}
	if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
		logger.Debug("Failed to clear version record for %s: %v", logger.H(name), err)
	}
}

// PinnedVersionNeedsUpdate reports whether the pinned version differs from the one sofmani
// recorded. A package whose version was never recorded — installed by hand, or pinned after
// it was already installed — is reconciled to the pin on the next run.
func PinnedVersionNeedsUpdate(name string, pinned string) bool {
	installed := ReadInstalledVersion(name)
	if installed == pinned {
		logger.Debug("%s is pinned to %s and already at it", logger.H(name), pinned)
		return false
	}
	logger.Debug("%s is pinned to %s, recorded version is %q", logger.H(name), pinned, installed)
	return true
}
