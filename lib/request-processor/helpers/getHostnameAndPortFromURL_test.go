package helpers

import (
	"testing"
)

func TestGetHostnameAndPortFromURL(t *testing.T) {
	tests := []struct {
		url      string
		hostname string
		port     uint32
	}{
		{"httpsss", "", 0},
		{"http://localhost:4000", "localhost", 4000},
		{"http://localhost", "localhost", 80},
		{"https://localhost", "localhost", 443},
		{"ftp://localhost", "localhost", 0},
		{"https://test.com:8080/test?abc=123", "test.com", 8080},
	}

	for _, test := range tests {

		hostname, port := GetHostnameAndPortFromURL(test.url)

		if hostname != test.hostname {
			t.Errorf("For URL %s, expected hostname %s, but got %s", test.url, test.hostname, hostname)
		}
		if port != test.port {
			t.Errorf("For URL %s, expected port %d, but got %d", test.url, test.port, port)
		}
	}
}
