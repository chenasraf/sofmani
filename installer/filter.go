package installer

import (
	"strings"
)

func FilterIsMatch(filters []string, name string) bool {
	if len(filters) == 0 {
		return true
	}
	match := false
	for _, f := range filters {
		if strings.HasPrefix(f, "!") {
			continue
		}
		if strings.Contains(name, f) {
			match = true
			break
		}
	}
	for _, f := range filters {
		if strings.HasPrefix(f, "!") && strings.Contains(name, f[1:]) {
			return false
		}
	}
	return match
}
