package machine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

var (
	cachedMachineID string
	machineIDOnce   sync.Once
)

// GetMachineID returns a deterministic, unique machine identifier.
// The ID is stable across hostname changes and other system configuration changes.
// It uses platform-specific stable identifiers:
//   - macOS: IOPlatformUUID (hardware UUID)
//   - Linux: /etc/machine-id or /var/lib/dbus/machine-id
//   - Windows: Registry MachineGuid
//
// The raw identifier is hashed with SHA-256 and truncated to 16 characters
// for a more user-friendly format while maintaining uniqueness.
func GetMachineID() string {
	machineIDOnce.Do(func() {
		var rawID string
		var err error

		switch runtime.GOOS {
		case "darwin":
			rawID, err = getMachineIDDarwin()
		case "linux":
			rawID, err = getMachineIDLinux()
		case "windows":
			rawID, err = getMachineIDWindows()
		default:
			err = fmt.Errorf("unsupported platform: %s", runtime.GOOS)
		}

		if err != nil || rawID == "" {
			// Fallback: generate a persistent ID file in user config directory
			rawID, err = getOrCreateFallbackID()
			if err != nil {
				// Last resort: use a hash of available system info
				rawID = getFallbackSystemInfo()
			}
		}

		// Hash the raw ID for privacy and consistent format
		cachedMachineID = hashMachineID(rawID)
	})

	return cachedMachineID
}

// getMachineIDDarwin retrieves the hardware UUID on macOS using ioreg.
func getMachineIDDarwin() (string, error) {
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse the IOPlatformUUID from the output
	for line := range strings.SplitSeq(string(output), "\n") {
		if strings.Contains(line, "IOPlatformUUID") {
			// Format: "IOPlatformUUID" = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				uuid := strings.TrimSpace(parts[1])
				uuid = strings.Trim(uuid, "\"")
				return uuid, nil
			}
		}
	}

	return "", fmt.Errorf("IOPlatformUUID not found")
}

// getMachineIDLinux retrieves the machine ID on Linux from /etc/machine-id
// or /var/lib/dbus/machine-id as a fallback.
func getMachineIDLinux() (string, error) {
	// Try /etc/machine-id first (systemd)
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		id := strings.TrimSpace(string(data))
		if id != "" {
			return id, nil
		}
	}

	// Fallback to /var/lib/dbus/machine-id
	if data, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
		id := strings.TrimSpace(string(data))
		if id != "" {
			return id, nil
		}
	}

	return "", fmt.Errorf("machine-id not found")
}

// getMachineIDWindows retrieves the MachineGuid from the Windows registry.
func getMachineIDWindows() (string, error) {
	cmd := exec.Command("reg", "query",
		`HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography`,
		"/v", "MachineGuid")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse the MachineGuid from the output
	for line := range strings.SplitSeq(string(output), "\n") {
		if strings.Contains(line, "MachineGuid") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				return fields[len(fields)-1], nil
			}
		}
	}

	return "", fmt.Errorf("MachineGuid not found")
}

// getOrCreateFallbackID creates or reads a persistent machine ID file
// in the user's config directory as a fallback mechanism.
func getOrCreateFallbackID() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	sofmaniDir := configDir + "/sofmani"
	idFile := sofmaniDir + "/machine-id"

	// Try to read existing ID
	if data, err := os.ReadFile(idFile); err == nil {
		id := strings.TrimSpace(string(data))
		if id != "" {
			return id, nil
		}
	}

	// Create new ID based on available system info
	newID := generateFallbackID()

	// Ensure directory exists
	if err := os.MkdirAll(sofmaniDir, 0755); err != nil {
		return newID, nil // Return the ID even if we can't persist it
	}

	// Write the ID file
	if err := os.WriteFile(idFile, []byte(newID), 0644); err != nil {
		return newID, nil // Return the ID even if we can't persist it
	}

	return newID, nil
}

// generateFallbackID generates a unique ID using available system information.
func generateFallbackID() string {
	info := getFallbackSystemInfo()
	hash := sha256.Sum256([]byte(info))
	return hex.EncodeToString(hash[:])
}

// getFallbackSystemInfo gathers available system information for ID generation.
func getFallbackSystemInfo() string {
	var parts []string

	// Add hostname (may change, but better than nothing)
	if hostname, err := os.Hostname(); err == nil {
		parts = append(parts, hostname)
	}

	// Add user home directory path
	if home, err := os.UserHomeDir(); err == nil {
		parts = append(parts, home)
	}

	// Add config directory path
	if configDir, err := os.UserConfigDir(); err == nil {
		parts = append(parts, configDir)
	}

	// Add OS and architecture
	parts = append(parts, runtime.GOOS, runtime.GOARCH)

	return strings.Join(parts, "|")
}

// hashMachineID creates a SHA-256 hash of the raw machine ID
// and returns a truncated hex string (16 characters).
func hashMachineID(rawID string) string {
	hash := sha256.Sum256([]byte(rawID))
	fullHex := hex.EncodeToString(hash[:])
	// Return first 16 characters for a more user-friendly format
	return fullHex[:16]
}

// SetMachineID overrides the detected machine ID. This is primarily used for testing.
func SetMachineID(id string) {
	cachedMachineID = id
	// Reset the once so that future calls won't re-detect
	machineIDOnce.Do(func() {})
}

// ResetMachineID clears the cached machine ID. This is primarily used for testing.
func ResetMachineID() {
	cachedMachineID = ""
	machineIDOnce = sync.Once{}
}
