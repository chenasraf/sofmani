package installer

import (
	"strconv"
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/stretchr/testify/assert"
)

type MockInstaller struct {
	data             *appconfig.InstallerData
	isInstalled      bool
	needsUpdate      bool
	installError     error
	updateError      error
	checkInstall     error
	checkUpdate      error
	validationErrors []ValidationError
}

func (m *MockInstaller) GetData() *appconfig.InstallerData {
	return m.data
}

func (m *MockInstaller) CheckIsInstalled() (bool, error) {
	return m.isInstalled, m.checkInstall
}

func (m *MockInstaller) CheckNeedsUpdate() (bool, error) {
	return m.needsUpdate, m.checkUpdate
}

func (m *MockInstaller) Install() error {
	return m.installError
}

func (m *MockInstaller) Update() error {
	return m.updateError
}

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
