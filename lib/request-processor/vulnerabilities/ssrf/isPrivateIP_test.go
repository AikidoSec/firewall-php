package ssrf

import (
	"testing"
)

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"0177.00.00.01", true},
		{"0177.00.00.1", true},
		{"0177.00.00.0x1", true},
		{"0177.00.0.01", true},
		{"0177.00.0.1", true},
		{"0177.00.0.0x1", true},
		{"0177.00.0x0.01", true},
		{"0177.00.0x0.1", true},
		{"0177.00.0x0.0x1", true},
		{"0177.0.00.01", true},
		{"0177.0.00.1", true},
		{"0177.0.00.0x1", true},
		{"0177.0.0.01", true},
		{"0177.0.0.1", true},
		{"0177.0.0.0x1", true},
		{"0177.0.0x0.01", true},
		{"0177.0.0x0.1", true},
		{"0177.0.0x0.0x1", true},
		{"0177.0x0.00.01", true},
		{"0177.0x0.00.1", true},
		{"0177.0x0.00.0x1", true},
		{"0177.0x0.0.01", true},
		{"0177.0x0.0.1", true},
		{"0177.0x0.0.0x1", true},
		{"0177.0x0.0x0.01", true},
		{"0177.0x0.0x0.1", true},
		{"0177.0x0.0x0.0x1", true},
		{"127.00.00.01", true},
		{"127.00.00.1", true},
		{"127.00.00.0x1", true},
		{"127.00.0.01", true},
		{"127.00.0.1", true},
		{"127.00.0.0x1", true},
		{"127.00.0x0.01", true},
		{"127.00.0x0.1", true},
		{"127.00.0x0.0x1", true},
		{"127.0.00.01", true},
		{"127.0.00.1", true},
		{"127.0.00.0x1", true},
		{"127.0.0.01", true},
		{"127.0.0.1", true},
		{"127.0.0.0x1", true},
		{"127.0.0x0.01", true},
		{"127.0.0x0.1", true},
		{"127.0.0x0.0x1", true},
		{"127.0x0.00.01", true},
		{"127.0x0.00.1", true},
		{"127.0x0.00.0x1", true},
		{"127.0x0.0.01", true},
		{"127.0x0.0.1", true},
		{"127.0x0.0.0x1", true},
		{"127.0x0.0x0.01", true},
		{"127.0x0.0x0.1", true},
		{"127.0x0.0x0.0x1", true},
		{"0x7f.00.00.01", true},
		{"0x7f.00.00.1", true},
		{"0x7f.00.00.0x1", true},
		{"0x7f.00.0.01", true},
		{"0x7f.00.0.1", true},
		{"0x7f.00.0.0x1", true},
		{"0x7f.00.0x0.01", true},
		{"0x7f.00.0x0.1", true},
		{"0x7f.00.0x0.0x1", true},
		{"0x7f.0.00.01", true},
		{"0x7f.0.00.1", true},
		{"0x7f.0.00.0x1", true},
		{"0x7f.0.0.01", true},
		{"0x7f.0.0.1", true},
		{"0x7f.0.0.0x1", true},
		{"0x7f.0.0x0.01", true},
		{"0x7f.0.0x0.1", true},
		{"0x7f.0.0x0.0x1", true},
		{"0x7f.0x0.00.01", true},
		{"0x7f.0x0.00.1", true},
		{"0x7f.0x0.00.0x1", true},
		{"0x7f.0x0.0.01", true},
		{"0x7f.0x0.0.1", true},
		{"0x7f.0x0.0.0x1", true},
		{"0x7f.0x0.0x0.01", true},
		{"0x7f.0x0.0x0.1", true},
		{"0x7f.0x0.0x0.0x1", true},
		{"0177.00.01", true},
		{"0177.00.1", true},
		{"0177.00.0x1", true},
		{"0177.0.01", true},
		{"0177.0.1", true},
		{"0177.0.0x1", true},
		{"0177.0x0.01", true},
		{"0177.0x0.1", true},
		{"0177.0x0.0x1", true},
		{"127.00.01", true},
		{"127.00.1", true},
		{"127.00.0x1", true},
		{"127.0.01", true},
		{"127.0.1", true},
		{"127.0.0x1", true},
		{"127.0x0.01", true},
		{"127.0x0.1", true},
		{"127.0x0.0x1", true},
		{"0x7f.00.01", true},
		{"0x7f.00.1", true},
		{"0x7f.00.0x1", true},
		{"0x7f.0.01", true},
		{"0x7f.0.1", true},
		{"0x7f.0.0x1", true},
		{"0x7f.0x0.01", true},
		{"0x7f.0x0.1", true},
		{"0x7f.0x0.0x1", true},
		{"0177.01", true},
		{"0177.1", true},
		{"0177.0x1", true},
		{"127.01", true},
		{"127.1", true},
		{"127.0x1", true},
		{"0x7f.01", true},
		{"0x7f.1", true},
		{"0x7f.0x1", true},
		{"017700000001", true},
		{"2130706433", true},
		{"0x7f000001", true},
		{"0251.0376.0251.0376", true},
		{"10.0.0.1", true},              // Private IPv4 (10.0.0.0/8)
		{"192.168.1.1", true},           // Private IPv4 (192.168.0.0/16)
		{"8.8.8.8", false},              // Public IPv4 (Google DNS)
		{"169.254.1.1", true},           // Link-local IPv4 (169.254.0.0/16)
		{"fc00::1", true},               // Private IPv6 (fc00::/7)
		{"fe80::1", true},               // Link-local IPv6 (fe80::/10)
		{"::1", true},                   // Loopback IPv6 (::1/128)
		{"::ffff:127.0.0.1", true},      // IPv4-mapped IPv6 (::ffff:127.0.0.1/128)
		{"::", true},                    // Unspecified IPv6 (::/128)
		{"2001:4860:4860::8888", false}, // Public IPv6 (Google DNS)
		{"240.0.0.1", true},             // Reserved IPv4 (240.0.0.0/4)
		{"255.255.255.255", true},       // Broadcast IPv4 (255.255.255.255/32)
	}

	for _, test := range tests {
		result := isPrivateIP(test.ip)
		if result != test.expected {
			t.Errorf("For IP '%s', expected %v but got %v", test.ip, test.expected, result)
		}
	}
}
