package summary

import (
	"strings"

	"github.com/chenasraf/sofmani/logger"
)

// Action represents the action taken for an installer.
type Action int

const (
	// ActionSkipped indicates the installer was skipped (platform/machine/filter mismatch).
	ActionSkipped Action = iota
	// ActionUpToDate indicates the software was already installed and up-to-date.
	ActionUpToDate
	// ActionInstalled indicates the software was newly installed.
	ActionInstalled
	// ActionUpgraded indicates the software was upgraded.
	ActionUpgraded
)

// InstallResult represents the result of running an installer.
type InstallResult struct {
	// Name is the name of the installer.
	Name string
	// Type is the installer type (e.g., "brew", "npm", "group").
	Type string
	// Action is the action that was taken.
	Action Action
	// Children contains results from nested installers (for group/manifest).
	Children []InstallResult
	// SkipSummaryInstall indicates whether to exclude this from install summary.
	SkipSummaryInstall bool
	// SkipSummaryUpdate indicates whether to exclude this from update summary.
	SkipSummaryUpdate bool
}

// Summary collects installation results for final reporting.
type Summary struct {
	results []InstallResult
}

// NewSummary creates a new Summary instance.
func NewSummary() *Summary {
	return &Summary{
		results: []InstallResult{},
	}
}

// Add adds an installation result to the summary.
func (s *Summary) Add(result InstallResult) {
	s.results = append(s.results, result)
}

// Print outputs the summary to the logger.
func (s *Summary) Print() {
	installed := s.collectByAction(ActionInstalled)
	upgraded := s.collectByAction(ActionUpgraded)

	hasInstalled := len(installed) > 0
	hasUpgraded := len(upgraded) > 0

	if !hasInstalled && !hasUpgraded {
		logger.Info("Summary: Nothing new to install or upgrade")
		return
	}

	logger.Info("Summary:")

	if hasInstalled {
		logger.Info("  Installed:")
		for _, r := range installed {
			s.printResult(r, 2)
		}
	}

	if hasUpgraded {
		logger.Info("  Upgraded:")
		for _, r := range upgraded {
			s.printResult(r, 2)
		}
	}
}

// collectByAction returns all results (including nested) that match the given action.
func (s *Summary) collectByAction(action Action) []InstallResult {
	var results []InstallResult
	for _, r := range s.results {
		collected := collectResultsByAction(r, action)
		results = append(results, collected...)
	}
	return results
}

// isContainerType returns true if the installer type is a container (group/manifest).
func isContainerType(installerType string) bool {
	return installerType == "group" || installerType == "manifest"
}

// shouldSkipSummary checks if a result should be skipped from summary based on action.
func shouldSkipSummary(r InstallResult, action Action) bool {
	if action == ActionInstalled && r.SkipSummaryInstall {
		return true
	}
	if action == ActionUpgraded && r.SkipSummaryUpdate {
		return true
	}
	return false
}

// collectResultsByAction recursively collects results matching the action.
// For group/manifest installers, it returns the parent with filtered children.
func collectResultsByAction(r InstallResult, action Action) []InstallResult {
	var results []InstallResult

	// Check if this result should be skipped from summary
	if shouldSkipSummary(r, action) {
		return results
	}

	if isContainerType(r.Type) {
		// For container types (groups/manifests), only include if children match
		if hasChildrenWithAction(r.Children, action) {
			filtered := InstallResult{
				Name:     r.Name,
				Type:     r.Type,
				Action:   r.Action,
				Children: filterChildrenByAction(r.Children, action),
			}
			results = append(results, filtered)
		}
	} else {
		// For leaf installers, include if action matches directly
		if r.Action == action {
			results = append(results, InstallResult{
				Name:   r.Name,
				Type:   r.Type,
				Action: r.Action,
			})
		}
	}

	return results
}

// filterChildrenByAction filters children to only include those matching the action.
func filterChildrenByAction(children []InstallResult, action Action) []InstallResult {
	var filtered []InstallResult
	for _, child := range children {
		// Skip if this child should be excluded from summary
		if shouldSkipSummary(child, action) {
			continue
		}

		if isContainerType(child.Type) {
			// For containers, only include if they have matching children
			if hasChildrenWithAction(child.Children, action) {
				filtered = append(filtered, InstallResult{
					Name:     child.Name,
					Type:     child.Type,
					Action:   child.Action,
					Children: filterChildrenByAction(child.Children, action),
				})
			}
		} else {
			// For leaf installers, include if action matches
			if child.Action == action {
				filtered = append(filtered, InstallResult{
					Name:   child.Name,
					Type:   child.Type,
					Action: child.Action,
				})
			}
		}
	}
	return filtered
}

// hasChildrenWithAction checks if any children (recursively) match the action.
func hasChildrenWithAction(children []InstallResult, action Action) bool {
	for _, child := range children {
		// Skip children that should be excluded from summary
		if shouldSkipSummary(child, action) {
			continue
		}
		if child.Action == action {
			return true
		}
		if hasChildrenWithAction(child.Children, action) {
			return true
		}
	}
	return false
}

// printResult prints a single result with the given indentation level.
func (s *Summary) printResult(r InstallResult, indent int) {
	prefix := strings.Repeat("  ", indent)
	logger.Info("%s- %s: %s", prefix, r.Type, r.Name)

	for _, child := range r.Children {
		s.printResult(child, indent+1)
	}
}
