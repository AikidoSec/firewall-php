package ssrf

import (
	. "main/aikido_types"
	"main/context"
	"main/helpers"
	"main/instance"
	"main/utils"
)

// allowedDomains defines the list of domains that are allowed for outgoing HTTP requests
// from PHP stream wrappers (file_get_contents, fopen, file) when the URL is user-controlled.
// This prevents SSRF attacks via redirects to internal addresses, which cannot be detected
// for stream wrappers (unlike curl which provides the effective URL after redirects).
var allowedDomains = []string{"example.com"} // add your allowed domains here

// CheckDomainAllowlistForStreamWrappers validates that the requested hostname is in the allowlist
// for PHP stream wrapper functions that cannot detect redirects.
// This is the primary defense against redirect-based SSRF for stream wrappers.
// Returns an InterceptorResult if the domain is not allowed, nil otherwise.
func CheckDomainAllowlistForStreamWrappers(instance *instance.RequestProcessorInstance, hostname string, port uint32, operation string) *utils.InterceptorResult {
	// Only apply this check to stream wrapper functions that cannot detect redirects
	// curl is handled separately and can detect redirects via CURLINFO_EFFECTIVE_URL
	isStreamWrapper := operation == "file" || operation == "file_get_contents" || operation == "fopen"
	if !isStreamWrapper {
		return nil
	}

	// Check if hostname was provided via user input
	hostnameFoundInUserInput := false
	var matchedSource string
	var matchedPath string
	var matchedPayload string

	for _, source := range context.SOURCES {
		mapss := source.CacheGet(instance)
		for str, path := range mapss {
			// Use the same hostname matching logic as findHostnameInUserInput
			if findHostnameInUserInput(str, hostname, port) {
				hostnameFoundInUserInput = true
				matchedSource = source.Name
				matchedPath = path
				matchedPayload = str
				break
			}
		}
		if hostnameFoundInUserInput {
			break
		}
	}

	// If hostname was not found in user input, no SSRF risk
	if !hostnameFoundInUserInput {
		return nil
	}

	// Normalize the hostname for comparison
	normalizedHostname := helpers.NormalizeHostname(hostname)

	// Check if domain is in the allowlist (exact match only)
	for _, allowedDomain := range allowedDomains {
		normalizedAllowed := helpers.NormalizeHostname(allowedDomain)
		if normalizedHostname == normalizedAllowed {
			// Domain is explicitly allowed
			return nil
		}
	}

	// Check if domain is explicitly allowed in cloud configuration
	var server *ServerData
	if instance != nil {
		server = instance.GetCurrentServer()
	}
	if server != nil {
		server.CloudConfigMutex.Lock()
		block, found := server.CloudConfig.OutboundDomains[normalizedHostname]
		server.CloudConfigMutex.Unlock()

		if found && !block {
			// Domain is explicitly allowed in cloud config
			return nil
		}
	}

	// Domain is not in allowlist - block the request to prevent redirect-based SSRF
	return &utils.InterceptorResult{
		Operation:     operation,
		Kind:          utils.Ssrf,
		Source:        matchedSource,
		PathToPayload: matchedPath,
		Metadata: map[string]string{
			"hostname":                hostname,
			"blockedByAllowlist":      "true",
			"reason":                  "redirect_protection",
			"streamWrapperFunction":   operation,
		},
		Payload: matchedPayload,
	}
}
