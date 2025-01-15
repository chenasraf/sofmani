package installer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterIsMatch(t *testing.T) {
	tests := []struct {
		name     string
		filters  []string
		input    string
		expected bool
	}{
		{"No filters", []string{}, "test", true},
		{"Match found", []string{"test"}, "test", true},
		{"No match", []string{"example"}, "test", false},
		{"Negation filter", []string{"!test"}, "test", false},
		{"Negation filter with match", []string{"example", "!test"}, "test", false},
		{"Negation filter without match", []string{"example", "!test"}, "example", true},
		{"Negation on included filter", []string{"example", "!example-test"}, "example-test", false},
		{"Partial match", []string{"config"}, "example-config", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterIsMatch(tt.filters, tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
