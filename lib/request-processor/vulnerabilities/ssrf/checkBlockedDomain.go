package ssrf

import (
	"main/globals"
	"main/helpers"
	"strings"
)

// IsBlockOutboundConnection checks if an outbound request to a hostname should be blocked
// based on the cloud configuration for blocked/allowed domains
func IsBlockOutboundConnection(hostname string) bool {
	server := globals.GetCurrentServer()
	if server == nil {
		return false
	}

	trimmedHostname := helpers.TrimInvisible(hostname)
	// Normalize hostname to lowercase for case-insensitive comparison
	normalizedHostname := strings.ToLower(trimmedHostname)

	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()

	// Check if hostname is in the outbound domains list
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
