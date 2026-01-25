package machine

// Machines defines which machines a configuration applies to.
type Machines struct {
	// Only specifies a list of machine IDs or aliases where the configuration should apply.
	Only *[]string `json:"only"   yaml:"only"`
	// Except specifies a list of machine IDs or aliases where the configuration should not apply.
	Except *[]string `json:"except" yaml:"except"`
}

// GetShouldRunOnMachine determines if a configuration should run on the current machine
// based on the Only and Except fields of the Machines struct.
// The aliases parameter is a map of friendly names to machine IDs. When checking,
// aliases are resolved first, falling back to treating the value as a literal machine ID.
func (m *Machines) GetShouldRunOnMachine(machineID string, aliases map[string]string) bool {
	if m == nil {
		return true
	}

	if m.Only != nil {
		return containsMachineID(*m.Only, machineID, aliases)
	}
	if m.Except != nil {
		return !containsMachineID(*m.Except, machineID, aliases)
	}
	return true
}

// containsMachineID checks if the machine ID is in the list, resolving aliases first.
func containsMachineID(list []string, machineID string, aliases map[string]string) bool {
	for _, entry := range list {
		// First, try to resolve as an alias
		if aliases != nil {
			if resolvedID, ok := aliases[entry]; ok {
				if resolvedID == machineID {
					return true
				}
				continue
			}
		}
		// Fall back to treating entry as a literal machine ID
		if entry == machineID {
			return true
		}
	}
	return false
}
