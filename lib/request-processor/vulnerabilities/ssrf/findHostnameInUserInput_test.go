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
		{"http://[0:0:0:0:0:ffff:127.0.0.1]/thefile", "0:0:0:0:0:ffff:127.0.0.1", 80, true},
		{"http://[0000:0000:0000:0000:0000:0000:0000:0001]:4000", "0000:0000:0000:0000:0000:0000:0000:0001", 4000, true},
		{"http://127.0.0.1:8080#\\@127.2.2.2:80/ ", "127.0.0.1", 8080, true},
		{"http://1.1.1.1 &@127.0.0.1:4000# @3.3.3.3/", "127.0.0.1", 4000, true},
		{"http://127.1.1.1:4000:\\@@127.0.0.1:8080/", "127.0.0.1", 8080, true},
		{"http://127.1.1.1:4000\\@127.0.0.1:8080/", "127.0.0.1", 8080, true},
		{"http://%31%32%37.%30.%30.%31:4000", "127.0.0.1", 4000, true},
		{"http://%30:4000", "0", 4000, true},
		{"http://127%2E0%2E0%2E1:4000", "127.0.0.1", 4000, true},
		{"http://[::ffff:127.0.0.1]:4000", "::ffff:127.0.0.1", 4000, true},
		{"http://[0:0:0:0:0:0:0:1]:4000", "::1", 4000, true},
		{"http://[::]:4000", "::", 4000, true},
		{"http://[0000:0000:0000:0000:0000:0000:0000:0001]:4000", "::1", 4000, true},
		{"http://[::1]:4000", "::1", 4000, true},
		{"http://[0:0::1]:4000", "::1", 4000, true},
		{"https://m%C3%BCnchen.de", "münchen.de", 0, true},
		{"https://münchen.de", "xn--mnchen-3ya.de", 0, true},
		{"https://xn--mnchen-3ya.de", "münchen.de", 0, true},
		{"hTTps://lOcalhosT:8081", "Localhost", 8081, true},
		{"MÜNCHEN.DE", "münchen.de", 0, true},
		{"HTTP://localhost", "loCalhost", 0, true},
		{"http://LOCALHOST", "loCalhOst", 0, true},
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
