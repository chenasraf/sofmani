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
		if strings.Contains(f, name) {
			match = true
			break
		}
	}
	for _, f := range filters {
		if strings.HasPrefix(f, "!") && strings.Contains(f[1:], name) {
			return false
		}
	}
	return match
}
