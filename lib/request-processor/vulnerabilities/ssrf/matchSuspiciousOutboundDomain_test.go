package ssrf

import (
	"main/aikido_types"
	"testing"
)

func TestMatchSuspiciousOutboundDomain(t *testing.T) {
	server := &aikido_types.ServerData{}
	if got := matchSuspiciousOutboundDomain(server, "oast.live"); got != "" {
		t.Fatalf("empty cloud list matched %q", got)
	}
	server.CloudConfig.SuspiciousOutboundDomains = map[string]bool{"oast.live": true, "lock.cmpxchg8b.com": true}

	for _, tc := range []struct{ host, want string }{
		{"oast.live", "oast.live"},
		{"a.b.oast.live", "oast.live"},
		{"Probe.OAST.Live.", "oast.live"},
		{"test.lock.cmpxchg8b.com", "lock.cmpxchg8b.com"},
		{"other.cmpxchg8b.com", ""},
		{"notoast.live", ""},
		{"oast.live.example.com", ""},
		{"", ""},
	} {
		if got := matchSuspiciousOutboundDomain(server, tc.host); got != tc.want {
			t.Errorf("matchSuspiciousOutboundDomain(%q) = %q, want %q", tc.host, got, tc.want)
		}
	}
}
