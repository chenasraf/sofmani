package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParsePrettyDuration parses a human-friendly duration string like "1d", "2w", "3m", "60s", "1h".
// Supported units: s (seconds), m (minutes), h (hours), d (days), w (weeks).
// Multiple components can be combined, e.g. "1d12h".
func ParsePrettyDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	var total time.Duration
	remaining := s

	for len(remaining) > 0 {
		// Find the first non-digit character
		i := 0
		for i < len(remaining) && remaining[i] >= '0' && remaining[i] <= '9' {
			i++
		}
		if i == 0 {
			return 0, fmt.Errorf("invalid duration %q: expected number", s)
		}
		if i >= len(remaining) {
			return 0, fmt.Errorf("invalid duration %q: missing unit", s)
		}

		num, err := strconv.ParseInt(remaining[:i], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", s, err)
		}

		unit := remaining[i]
		remaining = remaining[i+1:]

		switch unit {
		case 's':
			total += time.Duration(num) * time.Second
		case 'm':
			total += time.Duration(num) * time.Minute
		case 'h':
			total += time.Duration(num) * time.Hour
		case 'd':
			total += time.Duration(num) * 24 * time.Hour
		case 'w':
			total += time.Duration(num) * 7 * 24 * time.Hour
		default:
			return 0, fmt.Errorf("invalid duration %q: unknown unit %q", s, string(unit))
		}
	}

	return total, nil
}
