package installer

import (
	"errors"
	"fmt"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/summary"
	"github.com/samber/lo"
)

// Step is one entry of an install list: either a category header or an installer.
type Step struct {
	// Data is the entry as written in the config.
	Data *appconfig.InstallerData
	// Installer runs the entry. It is nil for a category header.
	Installer IInstaller
}

// IsCategory returns true if the step is a category header.
func (s Step) IsCategory() bool {
	return s.Installer == nil
}

// PrepareSteps builds and validates every entry of an install list. Category entries become
// header steps, and entries of an unsupported type are warned about and left out. Steps are only
// safe to run when no errors are returned.
func PrepareSteps(config *appconfig.AppConfig, entries []appconfig.InstallerData) ([]Step, []ValidationError) {
	steps := []Step{}
	problems := []ValidationError{}
	for idx := range entries {
		data := &entries[idx]
		if data.IsCategory() {
			steps = append(steps, Step{Data: data})
			continue
		}
		// Type-specific validation names the installer in its errors, so it only runs once the
		// installer has a name.
		if data.Name == nil || len(*data.Name) == 0 {
			problems = append(problems, ValidationError{
				FieldName:     "name",
				Message:       validationIsRequired(),
				InstallerName: fmt.Sprintf("%s installer #%d", data.Type, idx+1),
			})
			continue
		}
		inst, err := GetInstaller(config, data)
		if err != nil {
			problems = append(problems, ValidationError{FieldName: "type", Message: err.Error(), InstallerName: *data.Name})
			continue
		}
		if inst == nil {
			logger.Warn("Installer type %s is not supported, skipping", logger.H(string(data.Type)))
			continue
		}
		problems = append(problems, inst.Validate()...)
		steps = append(steps, Step{Data: data, Installer: inst})
	}
	return steps, problems
}

// RunSteps prepares and runs an install list nested inside another installer, printing its
// category headers and collecting the results of its installers.
func RunSteps(config *appconfig.AppConfig, entries []appconfig.InstallerData) ([]summary.InstallResult, error) {
	steps, problems := PrepareSteps(config, entries)
	if len(problems) > 0 {
		return nil, joinValidationErrors(problems)
	}
	results := []summary.InstallResult{}
	for _, step := range steps {
		if step.IsCategory() {
			logger.Category(*step.Data.Category, step.Data.Desc, logger.CategoryDisplayMode(config.GetCategoryDisplay()))
			continue
		}
		logger.Debug("Checking step %s", logger.H(*step.Data.Name))
		result, err := RunInstaller(config, step.Installer)
		if err != nil {
			return nil, fmt.Errorf("failed to run installer for step %s: %w", *step.Data.Name, err)
		}
		if result != nil {
			results = append(results, *result)
		}
	}
	return results, nil
}

// joinValidationErrors combines validation errors into a single error, one per line.
func joinValidationErrors(problems []ValidationError) error {
	return errors.Join(lo.Map(problems, func(e ValidationError, _ int) error { return e })...)
}
