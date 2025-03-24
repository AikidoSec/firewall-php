package helpers

import (
	"net"
)

// Taken from https://github.com/frenchbread/private-ip/blob/master/src/index.ts
var privateIPv4Ranges = []string{
	"0.0.0.0/8", // "This" network (RFC 1122)
	"10.0.0.0/8", // Private-Use Networks (RFC 1918)
	"100.64.0.0/10", // Shared Address Space (RFC 6598)
	"127.0.0.0/8", // Loopback (RFC 1122)
	"169.254.0.0/16", // Link Local (RFC 3927)
	"172.16.0.0/12", // Private-Use Networks (RFC 1918)
	"192.0.0.0/24", // IETF Protocol Assignments (RFC 5736)
	"192.0.2.0/24", // TEST-NET-1 (RFC 5737)
	"192.31.196.0/24", // AS112 Redirection Anycast (RFC 7535)
	"192.52.193.0/24", // Automatic Multicast Tunneling (RFC 7450)
	"192.88.99.0/24", // 6to4 Relay Anycast (RFC 3068)
	"192.168.0.0/16", // Private-Use Networks (RFC 1918)
	"192.175.48.0/24", // AS112 Redirection Anycast (RFC 7535)
	"198.18.0.0/15", // Network Interconnect Device Benchmark Testing (RFC 2544)
	"198.51.100.0/24", // TEST-NET-2 (RFC 5737)
	"203.0.113.0/24", // TEST-NET-3 (RFC 5737)
	"224.0.0.0/4", // Multicast (RFC 3171)
	"240.0.0.0/4", // Reserved for Future Use (RFC 1112)
	"255.255.255.255/32", // Limited Broadcast (RFC 919)
}

var privateIPv6Ranges = []string{
	"::/128", // Unspecified address (RFC 4291)
	"::1/128", // Loopback address (RFC 4291)
	"fc00::/7", // Unique local address (ULA) (RFC 4193)
	"fe80::/10", // Link-local address (LLA) (RFC 4291)
	"100::/64", // Discard prefix (RFC 6666)
	"2001:db8::/32", // Documentation prefix (RFC 3849)
	"3fff::/20", // Documentation prefix (RFC 9637)
}

// Parse the CIDR ranges into net.IPNet objects
var privateIPNets []*net.IPNet

func init() {
	// Add all the private IPv4 ranges to the list
	for _, cidr := range privateIPv4Ranges {
		_, ipNet, _ := net.ParseCIDR(cidr)
		privateIPNets = append(privateIPNets, ipNet)		
	}

	// Add all the private IPv6 ranges to the list
	for _, cidr := range privateIPv6Ranges {
		_, ipNet, _ := net.ParseCIDR(cidr)
		privateIPNets = append(privateIPNets, ipNet)
	}

	// Add all the private IPv4 ranges as IPv4-mapped IPv6 addresses
	for _, cidr := range privateIPv4Ranges {
		ipv6Cidr := MapIPv4ToIPv6(cidr)
		_, ipNet, _ := net.ParseCIDR(ipv6Cidr)
		privateIPNets = append(privateIPNets, ipNet)
	}
}

// isPrivateIP checks if an IP address is within a private range.
func isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, ipNet := range privateIPNets {
		if ipNet.Contains(parsedIP) {
			return true
		}
	}

	return false
}

