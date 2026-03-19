package installer

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/machine"
	"github.com/chenasraf/sofmani/platform"
)

// TemplateVars holds variables available for template replacement.
type TemplateVars struct {
	// Tag is the full tag name (e.g., "v1.0.0").
	Tag string
	// Version is the version without the leading "v" (e.g., "1.0.0").
	Version string
	// Arch is the system architecture in Go format (e.g., "amd64", "arm64").
	Arch string
	// ArchAlias is the system architecture in common alias format (e.g., "x86_64", "arm64").
	ArchAlias string
	// ArchGnu is the system architecture in GNU/Linux format (e.g., "x86_64", "aarch64").
	ArchGnu string
	// OS is the current operating system (e.g., "macos", "linux", "windows").
	OS string
	// DeviceID is the unique machine identifier (truncated SHA-256 hash).
	DeviceID string
	// DeviceIDAlias is the friendly alias for the current machine, if one is defined in machine_aliases.
	DeviceIDAlias string
}

// legacyTokens maps old-style tokens to their TemplateVars field names.
var legacyTokens = map[string]string{
	"{tag}":             "Tag",
	"{version}":         "Version",
	"{arch}":            "Arch",
	"{arch_alias}":      "ArchAlias",
	"{arch_gnu}":        "ArchGnu",
	"{os}":              "OS",
	"{device_id}":       "DeviceID",
	"{device_id_alias}": "DeviceIDAlias",
}

// NewTemplateVars creates a new TemplateVars with the provided tag and current system info.
// The machineAliases parameter is a map of friendly names to machine IDs, used to resolve DeviceIDAlias.
func NewTemplateVars(tag string, machineAliases map[string]string) *TemplateVars {
	version, _ := strings.CutPrefix(tag, "v")
	deviceID := machine.GetMachineID()
	return &TemplateVars{
		Tag:           tag,
		Version:       version,
		Arch:          string(platform.GetArch()),
		ArchAlias:     platform.GetArchAlias(),
		ArchGnu:       platform.GetArchGnu(),
		OS:            string(platform.GetPlatform()),
		DeviceID:      deviceID,
		DeviceIDAlias: resolveDeviceAlias(deviceID, machineAliases),
	}
}

// resolveDeviceAlias returns the alias for the given machine ID by reverse-looking up the aliases map.
// Returns an empty string if no alias is found.
func resolveDeviceAlias(machineID string, aliases map[string]string) string {
	for alias, id := range aliases {
		if id == machineID {
			return alias
		}
	}
	return ""
}

// ApplyTemplate applies template variables to a string.
// It supports both Go template syntax (e.g., "{{ .Tag }}") and legacy token syntax (e.g., "{tag}").
// When legacy tokens are detected, a deprecation warning is logged at DEBUG level.
func ApplyTemplate(input string, vars *TemplateVars, installerName string) (string, error) {
	result := input

	// First, handle legacy token replacement with deprecation warnings
	result = applyLegacyTokens(result, vars, installerName)

	// Then, handle Go template syntax if present
	if strings.Contains(result, "{{") {
		tmpl, err := template.New("template").Parse(result)
		if err != nil {
			return "", err
		}
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, vars)
		if err != nil {
			return "", err
		}
		result = buf.String()
	}

	return result, nil
}

// applyLegacyTokens replaces legacy tokens and logs deprecation warnings.
func applyLegacyTokens(input string, vars *TemplateVars, installerName string) string {
	result := input

	for token, fieldName := range legacyTokens {
		if strings.Contains(result, token) {
			logger.Warn(
				"Deprecated: installer %q uses legacy token %q. Please migrate to Go template syntax: {{ .%s }}",
				installerName, token, fieldName,
			)
			var value string
			switch fieldName {
			case "Tag":
				value = vars.Tag
			case "Version":
				value = vars.Version
			case "Arch":
				value = vars.Arch
			case "ArchAlias":
				value = vars.ArchAlias
			case "ArchGnu":
				value = vars.ArchGnu
			case "OS":
				value = vars.OS
			case "DeviceID":
				value = vars.DeviceID
			case "DeviceIDAlias":
				value = vars.DeviceIDAlias
			}
			result = strings.ReplaceAll(result, token, value)
		}
	}

	return result
}
