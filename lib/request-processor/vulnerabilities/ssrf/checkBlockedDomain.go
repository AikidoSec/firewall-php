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
	block, found := server.CloudConfig.OutboundDomains[normalizedHostname]

	// If hostname is in the list with block=true, always block it
	if found && block {
		return true
	}

	// If blockNewOutgoingRequests is enabled
	if server.CloudConfig.BlockNewOutgoingRequests {
		// If hostname is in the list with block=false (allow), allow it
		if found && !block {
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
