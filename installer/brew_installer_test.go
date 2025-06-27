package installer

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
)

func newTestBrewInstaller(data *appconfig.InstallerData) *BrewInstaller {
	return &BrewInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Info: data,
	}
}

func TestBrewValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid: No tap specified (tap is optional)
	emptyData := &appconfig.InstallerData{
		Name: strPtr("test-brew-valid"),
		Type: appconfig.InstallerTypeBrew,
	}
	assertNoValidationErrors(t, newTestBrewInstaller(emptyData).Validate())

	// 🟢 Valid: Well-formed tap (contains slash, sufficient length)
	validData := &appconfig.InstallerData{
		Name: strPtr("test-brew-valid-tap"),
		Type: appconfig.InstallerTypeBrew,
		Opts: &map[string]any{"tap": "valid/tap"},
	}
	assertNoValidationErrors(t, newTestBrewInstaller(validData).Validate())

	// 🔴 Invalid: Tap is present but malformed (missing slash or too short)
	invalidData := &appconfig.InstallerData{
		Name: strPtr("test-brew-invalid-tap"),
		Type: appconfig.InstallerTypeBrew,
		Opts: &map[string]any{"tap": "invalid-tap"},
	}
	assertHasValidationErrors(t, newTestBrewInstaller(invalidData).Validate())
}

func simulateBrewCheck(input string, exitCode int) (logs string, updateNeeded bool, finalErr error) {
	logBuf := &bytes.Buffer{}
	needsUpdate, parseErr := parseBrewOutdatedOutput(strings.NewReader(input), logBuf)

	// Treat only negative/128+ as actual errors (or change as needed)
	if exitCode < 0 || exitCode >= 128 {
		return logBuf.String(), false, fmt.Errorf("brew exited with error code %d", exitCode)
	}

	if parseErr != nil {
		return logBuf.String(), false, parseErr
	}

	// Exit code >0 means updates are available — trust that
	if exitCode > 0 {
		return logBuf.String(), true, nil
	}

	// Exit code 0: trust the parsed JSON
	return logBuf.String(), needsUpdate, nil
}

func TestBrewNeedsUpdateWithExitCode(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		exitCode       int
		expectedLogs   string
		expectedUpdate bool
		expectErr      bool
	}{

		{
			name: "brew exit 1 (updates available)",
			input: `{
  "formulae": [],
  "casks": []
}`,
			exitCode:       1,
			expectedLogs:   "",
			expectedUpdate: true, // non-zero means updates
			expectErr:      false,
		},
		{
			name: "brew exit 0 (no updates)",
			input: `{
  "formulae": [],
  "casks": []
}`,
			exitCode:       0,
			expectedLogs:   "",
			expectedUpdate: false,
			expectErr:      false,
		},
		{
			name: "brew exit 1 with logs",
			input: `Auto-updating Homebrew...
{
  "formulae": [{ "name": "bash" }],
  "casks": []
}`,
			exitCode:       1,
			expectedLogs:   "Auto-updating Homebrew...\n",
			expectedUpdate: true,
			expectErr:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logs, update, err := simulateBrewCheck(tc.input, tc.exitCode)

			if tc.expectErr && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if update != tc.expectedUpdate {
				t.Errorf("unexpected update result: got %v, want %v", update, tc.expectedUpdate)
			}
			if logs != tc.expectedLogs {
				t.Errorf("unexpected logs:\nGot:\n%q\nWant:\n%q", logs, tc.expectedLogs)
			}
		})
	}
}
