package main

import (
	"main/attack"
	"main/context"
	"main/grpc"
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
func OnPreOutgoingRequest() string {
	defer context.ResetEventContext()

	if context.IsProtectionTurnedOff() {
		log.Infof("Protection is turned off -> will not run detection logic!")
		return ""
	}

	hostname, port := context.GetOutgoingRequestHostnameAndPort()
	operation := context.GetFunctionName()

	res := ssrf.CheckContextForSSRF(hostname, port, operation)
	if res != nil {
		return attack.ReportAttackDetected(res)
	}

	log.Info("[BEFORE] Got domain: ", hostname)
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
func OnPostOutgoingRequest() string {
	defer context.ResetEventContext()

	hostname, port := context.GetOutgoingRequestHostnameAndPort()
	effectiveHostname, effectivePort := context.GetOutgoingRequestEffectiveHostnameAndPort()
	resolvedIp := context.GetOutgoingRequestResolvedIp()
	if hostname == "" {
		return ""
	}

	log.Info("[AFTER] Got domain: ", hostname, " port: ", port)

	go grpc.OnDomain(hostname, port)
	if effectiveHostname != hostname {
		go grpc.OnDomain(effectiveHostname, effectivePort)
	}

	if context.IsProtectionTurnedOff() {
		log.Infof("Protection is turned off -> will not run detection logic!")
		return ""
	}

	res := ssrf.CheckResolvedIpForSSRF(resolvedIp)
	if effectiveHostname != hostname {
		log.Infof("EffectiveHostname \"%s\" is different than Hostname \"%s\"!", effectiveHostname, hostname)

		// After the request was made, the effective hostname is different that the initially requested one (redirects)
		if res == nil {
			// We double check here for SSRF on the effective hostname because some sinks might not provide the resolved IP address
			res = ssrf.CheckEffectiveHostnameForSSRF(effectiveHostname)
		}
	}

	if res != nil {
		/* Throw exception to PHP layer if blocking is enabled -> Response content is not returned to the PHP code */
		return attack.ReportAttackDetected(res)
	}
	return ""
}
