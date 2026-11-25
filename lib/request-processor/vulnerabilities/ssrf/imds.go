package ssrf

import (
	"main/helpers"
	"main/instance"
	"main/utils"
	"net"
	"strings"
)

var IMDSAddresses = map[string]struct{}{
	"169.254.169.254": {}, // AWS IMDS (IPv4)
	"fd00:ec2::254":   {}, // AWS IMDS (IPv6)
	"100.100.100.200": {}, // Alibaba Cloud IMDS
}

// Google cloud uses the same IP addresses for its metadata service
// However, you need to set specific headers to access it
// In order to not block legitimate requests, we should allow the IP addresses for Google Cloud
var trustedHosts = map[string]struct{}{
	"metadata.google.internal": {},
	"metadata.goog":            {},
}

func isIMDSIPAddress(ip string) bool {
	if _, exists := IMDSAddresses[ip]; exists {
		return true
	}
	return false
}

func isTrustedHostname(hostname string) bool {
	// Check if the hostname is in the trusted hosts map
	_, exists := trustedHosts[hostname]
	return exists
}

func FindIMDSIp(hostname string, resolvedIps []string) string {
	if isTrustedHostname(hostname) {
		return ""
	}

	for _, ip := range resolvedIps {
		ip = net.ParseIP(ip).String()
		if isIMDSIPAddress(ip) {
			return ip
		}
	}
	return ""
}

func CheckStoredSSRF(instance *instance.RequestProcessorInstance, hostname string, port uint32, operation, peerIP string) *utils.InterceptorResult {
	canonicalHost := strings.ToLower(strings.TrimSuffix(hostname, "."))
	if canonicalHost == "" || net.ParseIP(canonicalHost) != nil || isTrustedHostname(canonicalHost) {
		return nil
	}
	ips := []string{peerIP}
	if peerIP == "" {
		ips = helpers.ResolveHostname(instance, hostname)
	}
	ip := FindIMDSIp(canonicalHost, ips)
	if ip == "" {
		return nil
	}
	metadata := getMetadataForSSRFAttack(hostname, port)
	metadata["privateIP"] = ip
	return &utils.InterceptorResult{
		Kind:      utils.StoredSsrf,
		Operation: operation,
		Metadata:  metadata,
		Payload:   hostname,
	}
}
