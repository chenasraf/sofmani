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
	positives, negatives := splitFilters(filters)
	keep := len(positives) == 0 || matchesAny(installer, positives)
	return keep && !matchesAny(installer, negatives)
}

// ManifestFilters returns the filters the installers inside a manifest run with, and whether the
// manifest runs at all. A manifest that an exclusion filter matches does not run. One that an
// inclusion filter matches runs in full, minus the exclusions. Any other manifest runs with the
// same filters, so that `-f name` reaches installers declared inside manifests.
func ManifestFilters(installer IInstaller, filters []string) ([]string, bool) {
	if len(filters) == 0 {
		return nil, true
	}
	positives, negatives := splitFilters(filters)
	if matchesAny(installer, negatives) {
		return nil, false
	}
	if len(positives) > 0 && matchesAny(installer, positives) {
		return lo.Map(negatives, func(f string, _ int) string { return "!" + f }), true
	}
	return filters, true
}

// splitFilters separates inclusion filters from exclusion filters, stripping the `!` prefix of
// the latter.
func splitFilters(filters []string) ([]string, []string) {
	positives := lo.Filter(filters, func(filter string, i int) bool {
		return filter[0] != '!'
	})
	negatives := lo.FilterMap(filters, func(filter string, i int) (string, bool) {
		return filter[1:], filter[0] == '!'
	})
	return positives, negatives
}

// matchesAny checks if an installer matches at least one of the given filters.
func matchesAny(installer IInstaller, filters []string) bool {
	return lo.SomeBy(filters, func(f string) bool {
		return isFilteredIn(installer, f)
	})
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

	// Apply template variables to the enabled command
	cmd := *enabledCmd
	if vars := i.GetTemplateVars(); vars != nil {
		name := ""
		if i.GetData().Name != nil {
			name = *i.GetData().Name
		}
		if result, err := ApplyTemplate(cmd, vars, name); err == nil {
			cmd = result
		}
	}

	shell := utils.GetOSShell(i.GetData().EnvShell)
	args := utils.GetOSShellArgs(cmd)

	success, err := utils.RunCmdGetSuccess(i.GetData().Environ(), shell, args...)

	if err != nil {
		return false, err
	}

	return success, nil
}
