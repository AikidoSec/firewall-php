package ssrf

import (
	"main/aikido_types"
	"main/context"
	"main/globals"
	"main/helpers"
	"testing"

	"go4.org/netipx"
)

func setupTestServerForBlockedDomains(blockNewOutgoingRequests bool, outboundDomains map[string]bool, bypassedIps *netipx.IPSet, requestIp string) func() {
	// Normalize domain keys like config.go does at load time
	normalizedDomains := map[string]bool{}
	for domain, block := range outboundDomains {
		normalizedDomains[helpers.NormalizeHostname(domain)] = block
	}

	server := &aikido_types.ServerData{
		CloudConfig: aikido_types.CloudConfigData{
			BlockNewOutgoingRequests: blockNewOutgoingRequests,
			OutboundDomains:          normalizedDomains,
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
	outboundDomains := map[string]bool{
		"evil.com": true,
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
	outboundDomains := map[string]bool{
		"evil.com": true,
		"safe.com": false,
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
	outboundDomains := map[string]bool{
		"safe.com": false,
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
	outboundDomains := map[string]bool{
		"safe.com": false,
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
	outboundDomains := map[string]bool{
		"safe.com": false,
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
	outboundDomains := map[string]bool{
		"evil.com": true,
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
	outboundDomains := map[string]bool{}
	cleanup := setupTestServerForBlockedDomains(true, outboundDomains, nil, "")
	defer cleanup()

	isBlocked := IsBlockOutboundConnection("example.com")

	if !isBlocked {
		t.Error("Expected domain to be blocked when domains list is empty and blockNewOutgoingRequests is true")
	}
}

func TestIsBlockOutboundConnection_EmptyDomainsListWithBlockNewDisabled(t *testing.T) {
	// Test that all domains are allowed when the list is empty and blockNewOutgoingRequests is false
	outboundDomains := map[string]bool{}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	isBlocked := IsBlockOutboundConnection("example.com")

	if isBlocked {
		t.Error("Expected domain to be allowed when domains list is empty and blockNewOutgoingRequests is false")
	}
}

// Punycode bypass tests - ensure attackers cannot bypass blocked domain checks
// by using Punycode encoding (xn--...) instead of Unicode domain names

func TestIsBlockOutboundConnection_PunycodeBypass_BlockedUnicodeRequestedAsPunycode(t *testing.T) {
	// Test that a blocked Unicode domain is also blocked when requested using Punycode
	// The domain list contains "münchen.de" in Unicode, but attacker tries "xn--mnchen-3ya.de"
	outboundDomains := map[string]bool{
		"münchen.de": true,
	}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	// Attacker tries to bypass by using Punycode encoding
	isBlocked := IsBlockOutboundConnection("xn--mnchen-3ya.de")

	if !isBlocked {
		t.Error("Expected Punycode hostname to be blocked when Unicode equivalent is in blocked list")
	}
}

func TestIsBlockOutboundConnection_PunycodeBypass_BlockedPunycodeRequestedAsUnicode(t *testing.T) {
	// Test that a blocked Punycode domain is also blocked when requested using Unicode
	// The domain list contains "xn--mnchen-3ya.de" in Punycode, but attacker tries "münchen.de"
	outboundDomains := map[string]bool{
		"xn--mnchen-3ya.de": true,
	}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	// Attacker tries to bypass by using Unicode
	isBlocked := IsBlockOutboundConnection("münchen.de")

	if !isBlocked {
		t.Error("Expected Unicode hostname to be blocked when Punycode equivalent is in blocked list")
	}
}

func TestIsBlockOutboundConnection_PunycodeBypass_AllowedUnicodeRequestedAsPunycode(t *testing.T) {
	// Test that an allowed Unicode domain is also allowed when requested using Punycode
	outboundDomains := map[string]bool{
		"münchen.de": false,
	}
	cleanup := setupTestServerForBlockedDomains(true, outboundDomains, nil, "")
	defer cleanup()

	// Request using Punycode encoding
	isBlocked := IsBlockOutboundConnection("xn--mnchen-3ya.de")

	if isBlocked {
		t.Error("Expected Punycode hostname to be allowed when Unicode equivalent is in allow list")
	}
}

func TestIsBlockOutboundConnection_PunycodeBypass_AllowedPunycodeRequestedAsUnicode(t *testing.T) {
	// Test that an allowed Punycode domain is also allowed when requested using Unicode
	outboundDomains := map[string]bool{
		"xn--mnchen-3ya.de": false,
	}
	cleanup := setupTestServerForBlockedDomains(true, outboundDomains, nil, "")
	defer cleanup()

	// Request using Unicode
	isBlocked := IsBlockOutboundConnection("münchen.de")

	if isBlocked {
		t.Error("Expected Unicode hostname to be allowed when Punycode equivalent is in allow list")
	}
}

func TestIsBlockOutboundConnection_PunycodeBypass_MixedSubdomains(t *testing.T) {
	// Test with subdomains containing IDN
	outboundDomains := map[string]bool{
		"böse.evil.com": true,
	}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	// Attacker tries with Punycode subdomain (xn--bse-sna = böse)
	isBlocked := IsBlockOutboundConnection("xn--bse-sna.evil.com")

	if !isBlocked {
		t.Error("Expected Punycode subdomain to be blocked when Unicode equivalent is in blocked list")
	}
}

func TestIsBlockOutboundConnection_PunycodeBypass_RussianDomain(t *testing.T) {
	// Test with Cyrillic (Russian) domain - "москва.ru" (Moscow)
	outboundDomains := map[string]bool{
		"москва.ru": true,
	}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	// Attacker tries with Punycode
	isBlocked := IsBlockOutboundConnection("xn--80adxhks.ru")

	if !isBlocked {
		t.Error("Expected Punycode Cyrillic hostname to be blocked when Unicode equivalent is in blocked list")
	}
}

func TestIsBlockOutboundConnection_PunycodeBypass_ChineseDomain(t *testing.T) {
	// Test with Chinese domain - "中文.com"
	outboundDomains := map[string]bool{
		"中文.com": true,
	}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	// Attacker tries with Punycode
	isBlocked := IsBlockOutboundConnection("xn--fiq228c.com")

	if !isBlocked {
		t.Error("Expected Punycode Chinese hostname to be blocked when Unicode equivalent is in blocked list")
	}
}

func TestIsBlockOutboundConnection_PunycodeBypass_WithPortStripped(t *testing.T) {
	// Test that hostnames work correctly (port should already be stripped by caller)
	outboundDomains := map[string]bool{
		"münchen.de": true,
	}
	cleanup := setupTestServerForBlockedDomains(false, outboundDomains, nil, "")
	defer cleanup()

	// Just the hostname without port
	isBlocked := IsBlockOutboundConnection("xn--mnchen-3ya.de")

	if !isBlocked {
		t.Error("Expected Punycode hostname to be blocked")
	}
}
