package installer

import (
	"strings"

	"github.com/chenasraf/sofmani/utils"
	"github.com/samber/lo"
)

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

	keep := false
	if len(positives) == 0 {
		keep = true
	}

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

func isFilteredIn(installer IInstaller, filter string) bool {
	data := installer.GetData()
	if strings.HasPrefix(filter, "type:") {
		typeName := filter[len("type:"):]
		if strings.ToLower(string(data.Type)) == strings.ToLower(typeName) {
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

	err, success := utils.RunCmdGetSuccess(i.GetData().Environ(), shell, args...)

	if err != nil {
		return false, err
	}

	return success, nil
}
