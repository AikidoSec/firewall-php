package main

import (
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

All these checks first verify if the hostname was provided via user input.
Protects both curl and fopen wrapper functions (file_get_contents, etc...).
*/
func OnPreOutgoingRequest(inst *instance.RequestProcessorInstance) string {
	if context.IsEndpointProtectionTurnedOff(inst) {
		log.Infof(inst, "Protection is turned off -> will not run detection logic!")
		return ""
	}

	hostname, port := context.GetOutgoingRequestHostnameAndPort(inst)
	operation := context.GetFunctionName(inst)

	res := ssrf.CheckContextForSSRF(inst, hostname, port, operation)
	if res != nil {
		return attack.ReportAttackDetected(res, inst)
	}

	log.Info(inst, "[BEFORE] Got domain: ", hostname)
	//TODO: check if domain is blacklisted
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
func OnPostOutgoingRequest(inst *instance.RequestProcessorInstance) string {
	defer context.ResetEventContext(inst)

	hostname, port := context.GetOutgoingRequestHostnameAndPort(inst)
	effectiveHostname, effectivePort := context.GetOutgoingRequestEffectiveHostnameAndPort(inst)
	resolvedIp := context.GetOutgoingRequestResolvedIp(inst)
	if hostname == "" {
		return ""
	}

	log.Info(inst, "[AFTER] Got domain: ", hostname, " port: ", port)

	server := inst.GetCurrentServer()
	if server != nil {
		threadID := inst.GetThreadID()
		go grpc.OnDomain(threadID, server, hostname, port)
		if effectiveHostname != hostname {
			go grpc.OnDomain(threadID, server, effectiveHostname, effectivePort)
		}
	}

	if context.IsEndpointProtectionTurnedOff(inst) {
		log.Infof(inst, "Protection is turned off -> will not run detection logic!")
		return ""
	}

	if ssrf.IsRequestToItself(inst, effectiveHostname, effectivePort) {
		log.Infof(inst, "Request to itself detected -> will not run detection logic!")
		return ""
	}

	res := ssrf.CheckResolvedIpForSSRF(inst, resolvedIp)
	if effectiveHostname != hostname {
		log.Infof(inst, "EffectiveHostname \"%s\" is different than Hostname \"%s\"!", effectiveHostname, hostname)

		// After the request was made, the effective hostname is different that the initially requested one (redirects)
		if res == nil {
			// We double check here for SSRF on the effective hostname because some sinks might not provide the resolved IP address
			res = ssrf.CheckEffectiveHostnameForSSRF(inst, effectiveHostname)
		}
	}

	if res != nil {
		/* Throw exception to PHP layer if blocking is enabled -> Response content is not returned to the PHP code */
		return attack.ReportAttackDetected(res, inst)
	}
	return ""
}
