package ssrf

// This IP address is used by AWS EC2 instances to access the instance metadata service (IMDS)
// We should block any requests to these IP addresses
// This prevents STORED SSRF attacks that try to access the instance metadata service
var IMDSAddresses = map[string]struct{}{
	"169.254.169.254": {}, // AWS IMDS (IPv4)
	"fd00:ec2::254":   {}, // AWS IMDS (IPv6)
}

// Google cloud uses the same IP addresses for its metadata service
// However, you need to set specific headers to access it
// In order to not block legitimate requests, we should allow the IP addresses for Google Cloud
var trustedHosts = map[string]struct{}{
	"metadata.google.internal": {},
	"metadata.goog":            {},
}

func IsIMDSIPAddress(ip string) bool {
	if _, exists := IMDSAddresses[ip]; exists {
		return true
	}
	return false
}

func IsTrustedHostname(hostname string) bool {
	// Check if the hostname is in the trusted hosts map
	_, exists := trustedHosts[hostname]
	return exists
}

func ResolvesToIMDSIp(hostname string, resolvedIps string) {

}
