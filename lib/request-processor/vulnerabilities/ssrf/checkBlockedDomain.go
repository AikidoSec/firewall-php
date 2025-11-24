package ssrf

import (
	"main/globals"
	"main/helpers"
	"main/utils"
	"strconv"
	"strings"
)

// createBlockedDomainResult creates an InterceptorResult for a blocked domain
func createBlockedDomainResult(hostname string, port uint32, operation string) *utils.InterceptorResult {
	return &utils.InterceptorResult{
		Operation: operation,
		Kind:      utils.BlockedDomain,
		Source:    "outbound-request",
		Metadata: map[string]string{
			"hostname": hostname,
			"port":     strconv.Itoa(int(port)),
		},
		Payload: hostname,
	}
}

// CheckBlockedDomain checks if an outbound request to a hostname should be blocked
// based on the cloud configuration for blocked/allowed domains
func CheckBlockedDomain(hostname string, port uint32, operation string) *utils.InterceptorResult {
	server := globals.GetCurrentServer()
	if server == nil {
		return nil
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
		return createBlockedDomainResult(hostname, port, operation)
	}

	// If blockNewOutgoingRequests is enabled
	if server.CloudConfig.BlockNewOutgoingRequests {
		// If hostname has mode "allow", allow it
		if found && mode == "allow" {
			return nil
		}

		// If hostname is not in the list, block it
		if !found {
			return createBlockedDomainResult(hostname, port, operation)
		}
	}

	// Allow the connection
	return nil
}
