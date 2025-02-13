package ssrf

import (
	"main/context"
	"main/utils"
)

/* This is called before a request is made to check for SSRF and block the request (not execute it) if SSRF found */
func CheckContextForSSRF(hostname string, port uint32, operation string) *utils.InterceptorResult {
	for _, source := range context.SOURCES {
		mapss := source.CacheGet()

		for str, path := range mapss {

			if findHostnameInUserInput(str, hostname, port) {
				interceptorResult := utils.InterceptorResult{
					Operation:     operation,
					Kind:          utils.Ssrf,
					Source:        source.Name,
					PathToPayload: path,
					Metadata:      getMetadataForSSRFAttack(hostname, port),
					Payload:       str,
				}

				if containsPrivateIPAddress(hostname) {
					// Hostname was found in user input and is actually a private IP address (http://127.0.0.1) -> SSRF
					interceptorResult.Metadata["isPrivateIp"] = "true"
					return &interceptorResult
				}

				resolvedIpStatus := getResolvedIpStatusForHostname(hostname)
				if resolvedIpStatus != nil {
					interceptorResult.Metadata["resolvedIp"] = resolvedIpStatus.ip
					if resolvedIpStatus.isIMDS {
						// Hostname was found in user input and it resolves to an IMDS IP address -> SSRF
						interceptorResult.Metadata["isIMDSIp"] = "true"
					}
					if resolvedIpStatus.isPrivate {
						// Hostname was found in user input and it resolves to a private IP address -> SSRF
						interceptorResult.Metadata["isPrivateIp"] = "true"
					}
					return &interceptorResult
				}

				// Hostname matched in the user input but we did not managed to determine if it's a SSRF attack at this point.
				// Storing the matching information (interceptor result) in order to use it once the request completes,
				// as at that point we might have more information to determine if SSRF or not.
				context.EventContextSetCurrentSsrfInterceptorResult(&interceptorResult)
			}
		}
	}
	return nil
}

/* This is called after the request is made to check for SSRF in the effective hostname - hostname optained after redirects from the PHP library that made the request (curl) */
func CheckEffectiveHostnameForSSRF(effectiveHostname string) *utils.InterceptorResult {
	interceptorResult := context.GetCurrentSsrfInterceptorResult()
	if interceptorResult == nil {
		// The initially requested hostname was not found in the user input -> no SSRF
		return nil
	}

	interceptorResult.Metadata["effectiveHostname"] = effectiveHostname
	resolvedIpStatus := getResolvedIpStatusForHostname(effectiveHostname)
	if resolvedIpStatus != nil {
		interceptorResult.Metadata["resolvedIp"] = resolvedIpStatus.ip
		if resolvedIpStatus.isIMDS {
			// Hostname was found in user input and the effective hostname (after redirects) resolved to an IMDS IP address -> SSRF
			interceptorResult.Metadata["isIMDSIp"] = "true"
		}
		if resolvedIpStatus.isPrivate {
			// Hostname was found in user input and the effective hostname (after redirects) resolved to a private IP address -> SSRF
			interceptorResult.Metadata["isPrivateIp"] = "true"
		}
	}

	return nil
}

/* This is called after the request is made to check for SSRF in the resolvedIP - IP optained from the PHP library that made the request (curl) */
func CheckResolvedIpForSSRF(resolvedIp string) *utils.InterceptorResult {
	interceptorResult := context.GetCurrentSsrfInterceptorResult()
	if interceptorResult == nil {
		// The initially requested hostname was not found in the user input -> no SSRF
		return nil
	}

	if !isPrivateIP(resolvedIp) {
		// The resolved IP address is not private -> no SSRF
		return nil
	}

	// Hostname was found in user input and the resolved IP is a private IP address -> SSRF
	interceptorResult.Metadata["resolvedIp"] = resolvedIp
	interceptorResult.Metadata["isPrivateIp"] = "true"
	return interceptorResult
}
