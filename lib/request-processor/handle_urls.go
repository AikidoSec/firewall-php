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

func OnPreOutgoingRequest(instance *instance.RequestProcessorInstance) string {
	hostname, port := context.GetOutgoingRequestHostnameAndPort(instance)
	operation := context.GetFunctionName(instance)

	if !context.IsIpBypassed(instance) && ssrf.IsBlockedOutboundDomain(instance, hostname) {
		server := instance.GetCurrentServer()
		// A blocked request skips the post-request hook, so record the domain here.
		if server != nil {
			go grpc.OnDomain(server, instance.GetCurrentToken(), hostname, port)
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

// The request has already been sent. Blocking here can still keep curl_exec from returning the response.
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
		go grpc.OnDomain(server, instance.GetCurrentToken(), hostname, port)
		if effectiveHostname != hostname {
			go grpc.OnDomain(server, instance.GetCurrentToken(), effectiveHostname, effectivePort)
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

		if res == nil {
			// Some sinks omit the resolved IP, so check the redirect hostname too.
			res = ssrf.CheckEffectiveHostnameForSSRF(instance, effectiveHostname)
		}
	}

	if res != nil {
		return attack.ReportAttackDetected(res, instance)
	}
	return ""
}
