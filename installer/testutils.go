package installer

import (
	"strconv"
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/stretchr/testify/assert"
)

// MockInstaller is a mock implementation of the IInstaller interface for testing.
type MockInstaller struct {
	// data is the installer data for the mock installer.
	data *appconfig.InstallerData
	// isInstalled simulates whether the software is installed.
	isInstalled bool
	// needsUpdate simulates whether an update is needed.
	needsUpdate bool
	// installError simulates an error during installation.
	installError error
	// updateError simulates an error during update.
	updateError error
	// checkInstall simulates an error during the CheckIsInstalled check.
	checkInstall error
	// checkUpdate simulates an error during the CheckNeedsUpdate check.
	checkUpdate error
	// validationErrors simulates validation errors for the installer.
	validationErrors []ValidationError
}

// GetData returns the installer data for the mock installer.
func (m *MockInstaller) GetData() *appconfig.InstallerData {
	return m.data
}

// CheckIsInstalled simulates checking if the software is installed.
func (m *MockInstaller) CheckIsInstalled() (bool, error) {
	return m.isInstalled, m.checkInstall
}

// CheckNeedsUpdate simulates checking if an update is needed.
func (m *MockInstaller) CheckNeedsUpdate() (bool, error) {
	return m.needsUpdate, m.checkUpdate
}

// Install simulates installing the software.
func (m *MockInstaller) Install() error {
	return m.installError
}

// Update simulates updating the software.
func (m *MockInstaller) Update() error {
	return m.updateError
}

// Validate simulates validating the installer configuration.
func (m *MockInstaller) Validate() []ValidationError {
	return m.validationErrors
}

// strPtr returns a pointer to the given string.
func strPtr(s string) *string {
	return &s
}

// simulateBrewCheck simulates parsing output from `brew outdated --json`
// along with handling the exit code semantics.

// BrewExitError represents a failure exit code from a simulated `brew` call.
type BrewExitError struct {
	ExitCode int
}

func (e *BrewExitError) Error() string {
	return "brew exited with code " + strconv.Itoa(e.ExitCode)
}

// assertValidationError checks for a validation error on a specific field.
func assertValidationError(t *testing.T, errors []ValidationError, field string) {
	t.Helper()
	for _, err := range errors {
		if err.FieldName == field {
			return
		}
	}
	t.Errorf("expected validation error for field %q but not found", field)
}

// assertNoValidationErrors ensures that validation errors are empty.
func assertNoValidationErrors(t *testing.T, errors []ValidationError) {
	t.Helper()
	assert.Empty(t, errors)
}

func assertHasValidationErrors(t *testing.T, errors []ValidationError) {
	t.Helper()
	if len(errors) == 0 {
		t.Error("expected validation errors but got none")
	}
}
