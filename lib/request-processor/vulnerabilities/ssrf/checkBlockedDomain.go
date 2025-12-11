package ssrf

import (
	"main/globals"
	"main/helpers"
)

// IsBlockOutboundConnection checks if an outbound request to a hostname should be blocked
// based on the cloud configuration for blocked/allowed domains
func IsBlockOutboundConnection(hostname string) bool {
	server := globals.GetCurrentServer()
	if server == nil {
		return false
	}

	// Normalize the hostname (handles Punycode/IDN, invisible chars, case)
	normalizedHostname := helpers.NormalizeHostname(hostname)

	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()

	// Check if hostname is in the outbound domains list
	// Config keys are already normalized at load time (see grpc/config.go)
	mode, found := server.CloudConfig.OutboundDomains[normalizedHostname]

	// If hostname has mode "block", always block it
	if found && mode == "block" {
		return true
	}

	// If blockNewOutgoingRequests is enabled
	if server.CloudConfig.BlockNewOutgoingRequests {
		// If hostname has mode "allow", allow it
		if found && mode == "allow" {
			return false
		}

		// If hostname is not in the list, block it
		if !found {
			return true
		}
	}

	// Allow the connection
	return false
}
