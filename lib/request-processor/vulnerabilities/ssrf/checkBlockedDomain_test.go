package ssrf

import (
	"main/aikido_types"
	"main/context"
	"main/globals"
	"main/utils"
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

func TestCheckBlockedDomain_ExplicitlyBlockedDomain(t *testing.T) {
	// Test that explicitly blocked domains are always blocked
	outboundDomains := map[string]string{
		"evil.com": "block",
	}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	result := CheckBlockedDomain("evil.com", 443, "curl")

	if result == nil {
		t.Error("Expected blocked domain to be blocked, but it was allowed")
	}

	if result.Kind != utils.BlockedDomain {
		t.Errorf("Expected kind to be blocked_domain, got %s", result.Kind)
	}
}

func TestCheckBlockedDomain_ExplicitlyBlockedDomainRegardlessOfFlag(t *testing.T) {
	// Test that explicitly blocked domains are blocked even when blockNewOutgoingRequests is false
	outboundDomains := map[string]string{
		"evil.com": "block",
		"safe.com": "allow",
	}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	result := CheckBlockedDomain("evil.com", 443, "curl")

	if result == nil {
		t.Error("Expected blocked domain to be blocked regardless of blockNewOutgoingRequests flag")
	}
}

func TestCheckBlockedDomain_AllowedDomainWithBlockNewEnabled(t *testing.T) {
	// Test that allowed domains are allowed when blockNewOutgoingRequests is true
	outboundDomains := map[string]string{
		"safe.com": "allow",
	}
	cleanup := setupTestServerForBlockedDomains(true, outboundDomains, nil, "")
	defer cleanup()

	result := CheckBlockedDomain("safe.com", 443, "curl")

	if result != nil {
		t.Error("Expected allowed domain to be allowed when blockNewOutgoingRequests is true")
	}
}

func TestCheckBlockedDomain_NewDomainBlockedWhenFlagEnabled(t *testing.T) {
	// Test that new domains are blocked when blockNewOutgoingRequests is true
	outboundDomains := map[string]string{
		"safe.com": "allow",
	}
	cleanup := setupTestServerForBlockedDomains(true, outboundDomains, nil, "")
	defer cleanup()

	result := CheckBlockedDomain("unknown.com", 443, "curl")

	if result == nil {
		t.Error("Expected unknown domain to be blocked when blockNewOutgoingRequests is true")
	}

}

func TestCheckBlockedDomain_NewDomainAllowedWhenFlagDisabled(t *testing.T) {
	// Test that new domains are allowed when blockNewOutgoingRequests is false
	outboundDomains := map[string]string{
		"safe.com": "allow",
	}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	result := CheckBlockedDomain("unknown.com", 443, "curl")

	if result != nil {
		t.Error("Expected unknown domain to be allowed when blockNewOutgoingRequests is false")
	}
}

func TestCheckBlockedDomain_CaseInsensitiveHostname(t *testing.T) {
	// Test that hostname matching is case-insensitive
	outboundDomains := map[string]string{
		"evil.com": "block",
	}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	// Test with uppercase hostname
	result := CheckBlockedDomain("EVIL.COM", 443, "curl")

	if result == nil {
		t.Error("Expected uppercase hostname to be blocked (case-insensitive matching)")
	}

	// Test with mixed case
	result = CheckBlockedDomain("Evil.Com", 443, "curl")

	if result == nil {
		t.Error("Expected mixed case hostname to be blocked (case-insensitive matching)")
	}
}

func TestCheckBlockedDomain_NoServerReturnsNil(t *testing.T) {
	// Test that function returns nil when there's no server
	originalServer := globals.GetCurrentServer()
	globals.CurrentServer = nil
	defer func() {
		globals.CurrentServer = originalServer
	}()

	result := CheckBlockedDomain("evil.com", 443, "curl")

	if result != nil {
		t.Error("Expected nil result when there's no server")
	}
}

func TestCheckBlockedDomain_EmptyDomainsListWithBlockNewEnabled(t *testing.T) {
	// Test that all domains are blocked when the list is empty and blockNewOutgoingRequests is true
	outboundDomains := map[string]string{}
	cleanup := setupTestServerForBlockedDomains(true, outboundDomains, nil, "")
	defer cleanup()

	result := CheckBlockedDomain("example.com", 443, "curl")

	if result == nil {
		t.Error("Expected domain to be blocked when domains list is empty and blockNewOutgoingRequests is true")
	}
}

func TestCheckBlockedDomain_EmptyDomainsListWithBlockNewDisabled(t *testing.T) {
	// Test that all domains are allowed when the list is empty and blockNewOutgoingRequests is false
	outboundDomains := map[string]string{}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	result := CheckBlockedDomain("example.com", 443, "curl")

	if result != nil {
		t.Error("Expected domain to be allowed when domains list is empty and blockNewOutgoingRequests is false")
	}
}
