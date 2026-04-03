package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParsePrettyDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"60s", 60 * time.Second, false},
		{"5m", 5 * time.Minute, false},
		{"2h", 2 * time.Hour, false},
		{"1d", 24 * time.Hour, false},
		{"1w", 7 * 24 * time.Hour, false},
		{"3d", 3 * 24 * time.Hour, false},
		{"1d12h", 36 * time.Hour, false},
		{"1w2d", 9 * 24 * time.Hour, false},
		{"", 0, true},
		{"abc", 0, true},
		{"5", 0, true},
		{"5x", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParsePrettyDuration(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
