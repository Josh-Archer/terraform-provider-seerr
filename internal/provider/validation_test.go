package provider

import (
	"testing"
)

func TestURLRegex(t *testing.T) {
	regex := urlRegex()
	tests := []struct {
		url      string
		expected bool
	}{
		{"http://localhost", true},
		{"https://example.com", true},
		{"http://localhost:5055", true},
		{"https://example.com/seerr", true},
		{"http://localhost/", false},
		{"https://example.com/", false},
		{"http://localhost:5055/", false},
		{"https://example.com/seerr/", false},
		{"localhost", false},
		{"ftp://example.com", false},
		{"http://", false},
	}

	for _, test := range tests {
		if got := regex.MatchString(test.url); got != test.expected {
			t.Errorf("urlRegex().MatchString(%q) = %v; want %v", test.url, got, test.expected)
		}
	}
}
