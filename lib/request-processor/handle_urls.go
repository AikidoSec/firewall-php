package main

import (
	"fmt"
	"html"
	"main/attack"
	"main/context"
	"main/grpc"
	"main/instance"
	"main/log"
	ssrf "main/vulnerabilities/ssrf"
)

/*
	Defends agains:

- basic SSRF (local IP address used as hostname)
- direct SSRF attacks (hostname that resolves directly to a local IP address - does not go through redirects)
- direct IMDS SSRF attacks (hostname is an IMDS IP)
- blocked outbound domains (based on cloud configuration)

All these checks first verify if the hostname was provided via user input.
Protects both curl and fopen wrapper functions (file_get_contents, etc...).
*/
func OnPreOutgoingRequest(instance *instance.RequestProcessorInstance) string {
	hostname, port := context.GetOutgoingRequestHostnameAndPort(instance)
	operation := context.GetFunctionName(instance)

	// Check if the domain is blocked based on cloud configuration
	if !context.IsIpBypassed(instance) && ssrf.IsBlockedOutboundDomainWithInst(instance, hostname) {
		server := instance.GetCurrentServer()
		// Blocked domains should also be reported to the agent.
		if server != nil {
			go grpc.OnDomain(instance.GetThreadID(), server, instance.GetCurrentToken(), hostname, port)
		}
		message := fmt.Sprintf("Aikido firewall has blocked an outbound connection: %s(...) to %s", operation, html.EscapeString(hostname))
		return attack.GetThrowAction(message, 500)
	}

	if context.IsEndpointProtectionTurnedOff(instance) {
		log.Infof(instance, "Protection is turned off -> will not run detection logic!")
		return ""
	}

	res := ssrf.CheckContextForSSRF(instance, hostname, port, operation)
	if res != nil {
		return attack.ReportAttackDetected(res, instance)
	}

	log.Info(instance, "[BEFORE] Got domain: ", hostname)
	return ""
}

/*
	This function acts as a last resort to protect against SSRF.
	If we didn't have enough info to stop the SSRF attack before the request was made,
	we attempt to block it after the request was made.
	If we detect SSRF here we throw an exception to the PHP layer and the response content
	of the request does NOT reach the PHP code, thus stopping the SSRF attack.
	If it's a PUT/POST request, it will actually go through, but an exception will be thrown to
	the PHP layer, thus downgrading it to blind SSRF.
	Defends agains:

- re-direct SSRF attacks (redirects lead to a hostname that resolves to a local IP address)
- re-direct IMDS SSRF attacks (redirects lead to a hostname that resolves to an IMDS IP address)

All these checks first verify if the hostname was provided via user input.
Protects curl.
*/
func OnPostOutgoingRequest(instance *instance.RequestProcessorInstance) string {
	defer context.ResetEventContext(instance)

	hostname, port := context.GetOutgoingRequestHostnameAndPort(instance)
	effectiveHostname, effectivePort := context.GetOutgoingRequestEffectiveHostnameAndPort(instance)
	resolvedIp := context.GetOutgoingRequestResolvedIp(instance)
	if hostname == "" {
		return ""
	}

	log.Info(instance, "[AFTER] Got domain: ", hostname, " port: ", port)

	server := instance.GetCurrentServer()
	if server != nil {
		go grpc.OnDomain(instance.GetThreadID(), server, instance.GetCurrentToken(), hostname, port)
		if effectiveHostname != hostname {
			go grpc.OnDomain(instance.GetThreadID(), server, instance.GetCurrentToken(), effectiveHostname, effectivePort)
		}
	}

	if context.IsEndpointProtectionTurnedOff(instance) {
		log.Infof(instance, "Protection is turned off -> will not run detection logic!")
		return ""
	}

	if ssrf.IsRequestToItself(instance, effectiveHostname, effectivePort) {
		log.Infof(instance, "Request to itself detected -> will not run detection logic!")
		return ""
	}

	res := ssrf.CheckResolvedIpForSSRF(instance, resolvedIp)
	if effectiveHostname != hostname {
		log.Infof(instance, "EffectiveHostname \"%s\" is different than Hostname \"%s\"!", effectiveHostname, hostname)

		// After the request was made, the effective hostname is different that the initially requested one (redirects)
		if res == nil {
			// We double check here for SSRF on the effective hostname because some sinks might not provide the resolved IP address
			res = ssrf.CheckEffectiveHostnameForSSRF(instance, effectiveHostname)
		}
	}

	if res != nil {
		/* Throw exception to PHP layer if blocking is enabled -> Response content is not returned to the PHP code */
		return attack.ReportAttackDetected(res, instance)
	}
	return ""
}
