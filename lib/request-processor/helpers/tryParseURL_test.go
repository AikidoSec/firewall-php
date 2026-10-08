package helpers

import (
	"net/url"
	"testing"
)

func TestTryParseURL_InvalidURL(t *testing.T) {
	input := "invalid"
	result := TryParseURL(input)
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestTryParseURL_ValidURL(t *testing.T) {
	input := "https://example.com"
	expected, _ := url.Parse(input)
	result := TryParseURL(input)

	if result == nil || result.String() != expected.String() {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestTryParseURL_IDNWithPort(t *testing.T) {
	// Test that IDN conversion works correctly when a port is present
	// This prevents SSRF bypass via Punycode IDN with explicit port
	input := "http://xn--mnchen-3ya.de:8080"
	result := TryParseURL(input)

	if result == nil {
		t.Errorf("Expected valid URL, got nil")
		return
	}

	// The hostname should be converted to Unicode
	expectedHostname := "münchen.de"
	if result.Hostname() != expectedHostname {
		t.Errorf("Expected hostname %v, got %v", expectedHostname, result.Hostname())
	}

	// The port should be preserved
	expectedPort := "8080"
	if result.Port() != expectedPort {
		t.Errorf("Expected port %v, got %v", expectedPort, result.Port())
	}

	// The Host field should contain both Unicode hostname and port
	expectedHost := "münchen.de:8080"
	if result.Host != expectedHost {
		t.Errorf("Expected Host %v, got %v", expectedHost, result.Host)
	}
}

func TestTryParseURL_IDNWithoutPort(t *testing.T) {
	// Test that IDN conversion still works for URLs without explicit port
	input := "http://xn--mnchen-3ya.de"
	result := TryParseURL(input)

	if result == nil {
		t.Errorf("Expected valid URL, got nil")
		return
	}

	// The hostname should be converted to Unicode
	expectedHostname := "münchen.de"
	if result.Hostname() != expectedHostname {
		t.Errorf("Expected hostname %v, got %v", expectedHostname, result.Hostname())
	}

	// The Host field should contain only the Unicode hostname
	if result.Host != expectedHostname {
		t.Errorf("Expected Host %v, got %v", expectedHostname, result.Host)
	}
}
