package ssrf

import (
	"main/aikido_types"
	"main/context"
	"main/globals"
	"testing"

	"go4.org/netipx"
)

func setupTestServerForBlockedDomains(blockNewOutgoingRequests bool, outboundDomains map[string]string, bypassedIps *netipx.IPSet, requestIp string) func() {
	server := &aikido_types.ServerData{
		CloudConfig: aikido_types.CloudConfigData{
			BlockNewOutgoingRequests: blockNewOutgoingRequests,
			OutboundDomains:          outboundDomains,
			BypassedIps:              bypassedIps,
		},
	}

	// Store original server and restore it later
	originalServer := globals.GetCurrentServer()
	globals.CurrentServer = server

	// Setup test context with request IP
	contextData := map[string]string{}
	if requestIp != "" {
		contextData["remoteAddress"] = requestIp
	}
	context.LoadForUnitTests(contextData)

	// Return cleanup function
	return func() {
		context.UnloadForUnitTests()
		globals.CurrentServer = originalServer
	}
}

func TestIsBlockOutboundConnection_ExplicitlyBlockedDomain(t *testing.T) {
	// Test that explicitly blocked domains are always blocked
	outboundDomains := map[string]string{
		"evil.com": "block",
	}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	isBlocked := IsBlockOutboundConnection("evil.com")

	if !isBlocked {
		t.Error("Expected blocked domain to be blocked, but it was allowed")
	}

}

func TestIsBlockOutboundConnection_ExplicitlyBlockedDomainRegardlessOfFlag(t *testing.T) {
	// Test that explicitly blocked domains are blocked even when blockNewOutgoingRequests is false
	outboundDomains := map[string]string{
		"evil.com": "block",
		"safe.com": "allow",
	}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	isBlocked := IsBlockOutboundConnection("evil.com")

	if !isBlocked {
		t.Error("Expected blocked domain to be blocked regardless of blockNewOutgoingRequests flag")
	}
}

func TestIsBlockOutboundConnection_AllowedDomainWithBlockNewEnabled(t *testing.T) {
	// Test that allowed domains are allowed when blockNewOutgoingRequests is true
	outboundDomains := map[string]string{
		"safe.com": "allow",
	}
	cleanup := setupTestServerForBlockedDomains(true, outboundDomains, nil, "")
	defer cleanup()

	isBlocked := IsBlockOutboundConnection("safe.com")

	if isBlocked {
		t.Error("Expected allowed domain to be allowed when blockNewOutgoingRequests is true")
	}
}

func TestIsBlockOutboundConnection_NewDomainBlockedWhenFlagEnabled(t *testing.T) {
	// Test that new domains are blocked when blockNewOutgoingRequests is true
	outboundDomains := map[string]string{
		"safe.com": "allow",
	}
	cleanup := setupTestServerForBlockedDomains(true, outboundDomains, nil, "")
	defer cleanup()

	isBlocked := IsBlockOutboundConnection("unknown.com")

	if !isBlocked {
		t.Error("Expected unknown domain to be blocked when blockNewOutgoingRequests is true")
	}

}

func TestIsBlockOutboundConnection_NewDomainAllowedWhenFlagDisabled(t *testing.T) {
	// Test that new domains are allowed when blockNewOutgoingRequests is false
	outboundDomains := map[string]string{
		"safe.com": "allow",
	}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	isBlocked := IsBlockOutboundConnection("unknown.com")

	if isBlocked {
		t.Error("Expected unknown domain to be allowed when blockNewOutgoingRequests is false")
	}
}

func TestIsBlockOutboundConnection_CaseInsensitiveHostname(t *testing.T) {
	// Test that hostname matching is case-insensitive
	outboundDomains := map[string]string{
		"evil.com": "block",
	}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	// Test with uppercase hostname
	isBlocked := IsBlockOutboundConnection("EVIL.COM")

	if !isBlocked {
		t.Error("Expected uppercase hostname to be blocked (case-insensitive matching)")
	}

	// Test with mixed case
	isBlocked = IsBlockOutboundConnection("Evil.Com")

	if !isBlocked {
		t.Error("Expected mixed case hostname to be blocked (case-insensitive matching)")
	}
}

func TestIsBlockOutboundConnection_NoServerReturnsNil(t *testing.T) {
	// Test that function returns nil when there's no server
	originalServer := globals.GetCurrentServer()
	globals.CurrentServer = nil
	defer func() {
		globals.CurrentServer = originalServer
	}()

	isBlocked := IsBlockOutboundConnection("evil.com")

	if isBlocked {
		t.Error("Expected nil isBlocked when there's no server")
	}
}

func TestIsBlockOutboundConnection_EmptyDomainsListWithBlockNewEnabled(t *testing.T) {
	// Test that all domains are blocked when the list is empty and blockNewOutgoingRequests is true
	outboundDomains := map[string]string{}
	cleanup := setupTestServerForBlockedDomains(true, outboundDomains, nil, "")
	defer cleanup()

	isBlocked := IsBlockOutboundConnection("example.com")

	if !isBlocked {
		t.Error("Expected domain to be blocked when domains list is empty and blockNewOutgoingRequests is true")
	}
}

func TestIsBlockOutboundConnection_EmptyDomainsListWithBlockNewDisabled(t *testing.T) {
	// Test that all domains are allowed when the list is empty and blockNewOutgoingRequests is false
	outboundDomains := map[string]string{}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	isBlocked := IsBlockOutboundConnection("example.com")

	if isBlocked {
		t.Error("Expected domain to be allowed when domains list is empty and blockNewOutgoingRequests is false")
	}
}
