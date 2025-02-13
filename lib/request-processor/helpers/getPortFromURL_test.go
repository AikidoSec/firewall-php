package helpers

import (
	"net/url"
	"testing"
)

func TestGetPortFromURL(t *testing.T) {
	tests := []struct {
		input    string
		expected uint32
	}{
		{"http://localhost:4000", 4000},
		{"http://localhost", 80},
		{"https://localhost", 443},
		{"ftp://localhost", 0},
		{"https://test.com:8080/test?abc=123", 8080},
	}

	for _, test := range tests {
		parsedURL, _ := url.Parse(test.input)
		result := GetPortFromURL(parsedURL)

		if result != test.expected {
			t.Errorf("For URL %s, expected port %d, but got %d", test.input, test.expected, result)
		}
	}
}
