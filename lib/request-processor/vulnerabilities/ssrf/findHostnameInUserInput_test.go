package ssrf

import (
	"testing"
)

func TestFindHostnameInUserInput(t *testing.T) {
	tests := []struct {
		userInput string
		hostname  string
		port      uint32
		expected  bool
	}{
		{"", "", 0, false},
		{"", "example.com", 0, false},
		{"http://example.com", "", 0, false},
		{"http://localhost", "localhost", 0, true},
		{"http://localhost", "localhost", 0, true},
		{"http://localhost/path", "localhost", 0, true},
		//{"http:/localhost", "localhost", 0, true},
		//{"http:localhost", "localhost", 0, true},
		//{"http:/localhost/path/path", "localhost", 0, true},
		{"localhost/path/path", "localhost", 0, true},
		{"ftp://localhost", "localhost", 0, true},
		{"localhost", "localhost", 0, true},
		{"http://", "localhost", 0, false},
		{"localhost", "localhost localhost", 0, false},
		{"http://169.254.169.254/latest/meta-data/", "169.254.169.254", 0, true},
		{"http://2130706433", "2130706433", 0, true},
		{"http://127.1", "127.1", 0, true},
		{"http://127.0.1", "127.0.1", 0, true},
		{"http://localhost", "localhost", 8080, false},
		{"http://localhost:8080", "localhost", 8080, true},
		{"http://localhost:8080", "localhost", 0, true},
		{"http://localhost:8080", "localhost", 4321, false},
		{"https://example.com", "example.com", 443, true},
		{"https://example.com", "google.com", 443, false},
		{"http://wikipedia.com", "wikipedia.com", 80, true},
		{"http://aikido.dev:9090/", "aikido.dev", 9090, true},
	}

	for _, test := range tests {
		result := findHostnameInUserInput(test.userInput, test.hostname, test.port)
		if result != test.expected {
			t.Errorf("For input '%s' and hostname '%s' with port %d, expected %v but got %v",
				test.userInput, test.hostname, test.port, test.expected, result)
		}
	}
}
