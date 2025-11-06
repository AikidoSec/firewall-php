package ssrf

import (
	"main/aikido_types"
	"main/context"
	"main/globals"
	"testing"
)

func setupTestContext(serverURL string, trustProxy bool) func() {
	// Setup a mock server with trust proxy setting
	server := &aikido_types.ServerData{
		AikidoConfig: aikido_types.AikidoConfigData{
			TrustProxy: trustProxy,
		},
	}

	// Store original server and restore it later
	originalServer := globals.GetCurrentServer()
	globals.CurrentServer = server

	// Use the proper test context loader
	context.LoadForUnitTests(map[string]string{
		"url": serverURL,
	})

	// Return cleanup function
	return func() {
		context.UnloadForUnitTests()
		globals.CurrentServer = originalServer
	}
}

func TestIsRequestToItself_ReturnsFalseIfHostnamesDifferent(t *testing.T) {
	cleanup := setupTestContext("http://aikido.dev:4000", true)
	defer cleanup()

	result := IsRequestToItself("google.com", 4000)
	if result != false {
		t.Errorf("Expected false when hostnames are different, got %v", result)
	}
}

func TestIsRequestToItself_ReturnsFalseIfHostnamesDifferentHTTPS(t *testing.T) {
	cleanup := setupTestContext("https://aikido.dev", true)
	defer cleanup()

	result := IsRequestToItself("google.com", 443)
	if result != false {
		t.Errorf("Expected false when hostnames are different (HTTPS), got %v", result)
	}
}

func TestIsRequestToItself_ReturnsFalseIfHostnamesDifferentWithCustomPort(t *testing.T) {
	cleanup := setupTestContext("https://aikido.dev:4000", true)
	defer cleanup()

	result := IsRequestToItself("google.com", 443)
	if result != false {
		t.Errorf("Expected false when hostnames are different (custom port), got %v", result)
	}
}

func TestIsRequestToItself_ReturnsTrueIfServerDoesRequestToItself(t *testing.T) {
	tests := []struct {
		serverURL        string
		outboundHostname string
		outboundPort     uint32
		description      string
	}{
		{
			serverURL:        "https://aikido.dev",
			outboundHostname: "aikido.dev",
			outboundPort:     443,
			description:      "HTTPS to HTTPS same port",
		},
		{
			serverURL:        "http://aikido.dev:4000",
			outboundHostname: "aikido.dev",
			outboundPort:     4000,
			description:      "HTTP custom port to same custom port",
		},
		{
			serverURL:        "http://aikido.dev",
			outboundHostname: "aikido.dev",
			outboundPort:     80,
			description:      "HTTP to HTTP default port",
		},
		{
			serverURL:        "https://aikido.dev:4000",
			outboundHostname: "aikido.dev",
			outboundPort:     4000,
			description:      "HTTPS custom port to same custom port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			cleanup := setupTestContext(tt.serverURL, true)
			defer cleanup()

			result := IsRequestToItself(tt.outboundHostname, tt.outboundPort)
			if result != true {
				t.Errorf("Expected true for %s, got %v", tt.description, result)
			}
		})
	}
}

func TestIsRequestToItself_ReturnsTrueForSpecialCaseHTTPHTTPS(t *testing.T) {
	tests := []struct {
		serverURL        string
		outboundHostname string
		outboundPort     uint32
		description      string
	}{
		{
			serverURL:        "http://aikido.dev",
			outboundHostname: "aikido.dev",
			outboundPort:     443,
			description:      "HTTP (port 80) to HTTPS (port 443)",
		},
		{
			serverURL:        "https://aikido.dev",
			outboundHostname: "aikido.dev",
			outboundPort:     80,
			description:      "HTTPS (port 443) to HTTP (port 80)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			cleanup := setupTestContext(tt.serverURL, true)
			defer cleanup()

			result := IsRequestToItself(tt.outboundHostname, tt.outboundPort)
			if result != true {
				t.Errorf("Expected true for special case %s, got %v", tt.description, result)
			}
		})
	}
}

func TestIsRequestToItself_ReturnsFalseIfTrustProxyIsFalse(t *testing.T) {
	tests := []struct {
		serverURL        string
		outboundHostname string
		outboundPort     uint32
		description      string
	}{
		{
			serverURL:        "https://aikido.dev",
			outboundHostname: "aikido.dev",
			outboundPort:     443,
			description:      "Same hostname and port but trust proxy disabled",
		},
		{
			serverURL:        "http://aikido.dev",
			outboundHostname: "aikido.dev",
			outboundPort:     80,
			description:      "Same hostname and port (HTTP) but trust proxy disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			cleanup := setupTestContext(tt.serverURL, false) // Trust proxy is false
			defer cleanup()

			result := IsRequestToItself(tt.outboundHostname, tt.outboundPort)
			if result != false {
				t.Errorf("Expected false when trust proxy is disabled for %s, got %v", tt.description, result)
			}
		})
	}
}
