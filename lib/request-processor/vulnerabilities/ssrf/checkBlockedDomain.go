package ssrf

import (
	. "main/aikido_types"
	"main/helpers"
	"main/instance"
)

// IsBlockedOutboundDomain checks if an outbound request to a hostname should be blocked
// based on the cloud configuration for blocked/allowed domains
func IsBlockedOutboundDomain(instance *instance.RequestProcessorInstance, hostname string) bool {
	var server *ServerData
	if instance != nil {
		server = instance.GetCurrentServer()
	}
	if server == nil {
		return false
	}

	// Normalize the hostname (handles Punycode/IDN, invisible chars, case)
	normalizedHostname := helpers.NormalizeHostname(hostname)

	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()

	// Check if hostname is in the outbound domains list
	block, found := server.CloudConfig.OutboundDomains[normalizedHostname]
	if !found {
		// If hostname is not in the list and blockNewOutgoingRequests is enabled, block it
		return server.CloudConfig.BlockNewOutgoingRequests
	}

	// if it is found, return the block value
	return block
}
