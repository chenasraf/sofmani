package installer

import (
	"strings"

	"github.com/chenasraf/sofmani/utils"
	"github.com/samber/lo"
)

// FilterInstaller determines whether an installer should be included based on a list of filters.
// Filters can be positive (e.g., "name") or negative (e.g., "!name").
// Filters can also target specific fields like type (e.g., "type:brew") or tags (e.g., "tag:database").
func FilterInstaller(installer IInstaller, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	positives := lo.Filter(filters, func(filter string, i int) bool {
		return filter[0] != '!'
	})

	negatives := lo.FilterMap(filters, func(filter string, i int) (string, bool) {
		return filter[1:], filter[0] == '!'
	})

	keep := len(positives) == 0

	for _, f := range positives {
		if isFilteredIn(installer, f) {
			keep = true
			break
		}
	}
	for _, f := range negatives {
		if isFilteredIn(installer, f) {
			keep = false
			break
		}
	}

	return keep
}

// isFilteredIn checks if a single installer matches a given filter.
func isFilteredIn(installer IInstaller, filter string) bool {
	data := installer.GetData()
	if strings.HasPrefix(filter, "type:") {
		typeName := filter[len("type:"):]
		if strings.EqualFold(string(data.Type), typeName) {
			return true
		}
	}
	if strings.HasPrefix(filter, "tag:") {
		tagName := filter[len("tag:"):]
		if lo.SomeBy(data.GetTagsList(), func(tag string) bool {
			return strings.ToLower(tag) == tagName
		}) {
			return true
		}
	}
	return strings.Contains(*data.Name, filter)
}

// InstallerIsEnabled checks if an installer is enabled.
// The "enabled" field in the installer data can be a boolean string ("true", "false") or a command.
// If it's a command, the installer is enabled if the command runs successfully (exit code 0).
func InstallerIsEnabled(i IInstaller) (bool, error) {
	enabledCmd := i.GetData().Enabled

	if enabledCmd == nil {
		return true, nil
	}

	if strings.ToLower(*enabledCmd) == "true" {
		return true, nil
	}

	if strings.ToLower(*enabledCmd) == "false" {
		return false, nil
	}

	shell := utils.GetOSShell(i.GetData().EnvShell)
	args := utils.GetOSShellArgs(*enabledCmd)

	success, err := utils.RunCmdGetSuccess(i.GetData().Environ(), shell, args...)

	if err != nil {
		return false, err
	}

	return success, nil
}
