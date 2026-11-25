package ssrf

import (
	"main/aikido_types"
	"main/helpers"
	"strings"
)

func matchSuspiciousOutboundDomain(server *aikido_types.ServerData, hostname string) string {
	if server == nil {
		return ""
	}
	server.CloudConfigMutex.Lock()
	domains := server.CloudConfig.SuspiciousOutboundDomains
	server.CloudConfigMutex.Unlock()
	if len(domains) == 0 {
		return ""
	}
	for host := helpers.NormalizeHostname(hostname); host != ""; {
		if domains[host] {
			return host
		}
		_, host, _ = strings.Cut(host, ".")
	}
	return ""
}
