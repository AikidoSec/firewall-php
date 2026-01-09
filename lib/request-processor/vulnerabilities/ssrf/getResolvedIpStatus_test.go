package ssrf

import (
	"testing"
)

func TestResolvedIpStatus(t *testing.T) {
	tests := []struct {
		hostname string
		expected *ResolvedIpStatus
	}{
		{"127.0.0.1", &ResolvedIpStatus{ip: "127.0.0.1", isPrivate: true, isIMDS: false}},
		{"169.254.169.254", &ResolvedIpStatus{ip: "169.254.169.254", isPrivate: true, isIMDS: true}},
	}

	for _, test := range tests {
		result := getResolvedIpStatusForHostname(nil, test.hostname)
		if result == nil {
			t.Errorf("For hostname '%s' expected DNS resolution to not fail", test.hostname)
			break
		}
		if result.ip != test.expected.ip {
			t.Errorf("For hostname '%s' expected ip %v but got ip %v",
				test.hostname, test.expected.ip, result.ip)
		}
		if result.isPrivate != test.expected.isPrivate {
			t.Errorf("For hostname '%s' expected isPrivate %v but got isPrivate %v",
				test.hostname, test.expected.isPrivate, result.isPrivate)
		}
		if result.isIMDS != test.expected.isIMDS {
			t.Errorf("For hostname '%s' expected isIMDS %v but got isIMDS %v",
				test.hostname, test.expected.isIMDS, result.isIMDS)
		}
	}
}
