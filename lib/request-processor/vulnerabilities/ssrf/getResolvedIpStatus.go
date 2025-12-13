package ssrf

import (
	"main/helpers"
	"main/instance"
)

type ResolvedIpStatus struct {
	ip        string
	isPrivate bool
	isIMDS    bool
}

/*
We do our own DNS resolution request here, in order to resolve the requested hostname to a list of IPs.
As we use the same OS call that PHP uses and we do this immediately before or after PHP makes it's own DNS resolution,
we expect that for most of the cases, the result will be already cached at the OS-level and we wouldn't incur a performance penaly.
We do our own DNS resolution, because we want to actually block potential SSRF attacks and we did not find any way to hook PHP's DNS
resolution calls.
*/
func getResolvedIpStatusForHostname(inst *instance.RequestProcessorInstance, hostname string) *ResolvedIpStatus {
	resolvedIps := helpers.ResolveHostname(inst, hostname)
	imdsIP := FindIMDSIp(hostname, resolvedIps)

	if imdsIP != "" {
		return &ResolvedIpStatus{ip: imdsIP, isPrivate: isPrivateIP(imdsIP), isIMDS: true}
	}

	privateIp := helpers.FindPrivateIp(resolvedIps)
	if privateIp != "" {
		return &ResolvedIpStatus{ip: privateIp, isPrivate: true, isIMDS: false}
	}
	return nil
}
