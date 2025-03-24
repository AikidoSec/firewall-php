package helpers

import (
	"testing"
)

func TestMapIPv4ToIPv6(t *testing.T) {
	tests := []struct {
		ip       string
		expected string
	}{
		{"127.0.0.0", "::ffff:127.0.0.0/128"},
		{"127.0.0.0/8", "::ffff:127.0.0.0/104"},
		{"10.0.0.0", "::ffff:10.0.0.0/128"},
		{"10.0.0.0/8", "::ffff:10.0.0.0/104"},
		{"10.0.0.1", "::ffff:10.0.0.1/128"},
		{"10.0.0.1/8", "::ffff:10.0.0.1/104"},
		{"192.168.0.0/16", "::ffff:192.168.0.0/112"},
		{"172.16.0.0/12", "::ffff:172.16.0.0/108"},
	}

	for _, test := range tests {
		result := MapIPv4ToIPv6(test.ip)
		if result != test.expected {
			t.Errorf("For IP '%s', expected %v but got %v", test.ip, test.expected, result)
		}
	}
}
