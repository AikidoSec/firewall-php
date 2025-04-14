package helpers

import (
	"testing"
)

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"10.0.0.1", true},              // Private IPv4 (10.0.0.0/8)
		{"192.168.1.1", true},           // Private IPv4 (192.168.0.0/16)
		{"8.8.8.8", false},              // Public IPv4 (Google DNS)
		{"127.0.0.1", true},             // Loopback IPv4 (127.0.0.0/8)
		{"169.254.1.1", true},           // Link-local IPv4 (169.254.0.0/16)
		{"fc00::1", true},               // Private IPv6 (fc00::/7)
		{"fe80::1", true},               // Link-local IPv6 (fe80::/10)
		{"::1", true},                   // Loopback IPv6 (::1/128)
		{"::ffff:127.0.0.1", true},      // IPv4-mapped IPv6 (::ffff:127.0.0.1/128)
		{"::", true},                    // Unspecified IPv6 (::/128)
		{"2001:4860:4860::8888", false}, // Public IPv6 (Google DNS)
		{"240.0.0.1", true},             // Reserved IPv4 (240.0.0.0/4)
		{"255.255.255.255", true},       // Broadcast IPv4 (255.255.255.255/32)
		{"::ffff:10.0.0.1", true},       // IPv4-mapped IPv6
		{"::ffff:192.168.2.1", true},    // IPv4-mapped IPv6
		{"::ffff:1.2.3.4", false},       // IPv4-mapped IPv6
		{"::ffff:192.145.45.45", false}, // IPv4-mapped IPv6
	}

	for _, test := range tests {
		result := isPrivateIP(test.ip)
		if result != test.expected {
			t.Errorf("For IP '%s', expected %v but got %v", test.ip, test.expected, result)
		}
	}
}
