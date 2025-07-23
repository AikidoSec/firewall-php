package ssrf

import (
	"net"
	"strconv"
	"strings"
)

// Taken from https://github.com/frenchbread/private-ip/blob/master/src/index.ts
var privateIPv4Ranges = []string{
	"0.0.0.0/8",
	"10.0.0.0/8",
	"100.64.0.0/10",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"192.31.196.0/24",
	"192.52.193.0/24",
	"192.88.99.0/24",
	"192.168.0.0/16",
	"192.175.48.0/24",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"240.0.0.0/4",
	"224.0.0.0/4",
	"255.255.255.255/32",
}

var privateIPv6Ranges = []string{
	"::/128",               // Unspecified address
	"::1/128",              // Loopback address
	"fc00::/7",             // Unique local address (ULA)
	"fe80::/10",            // Link-local address (LLA)
	"::ffff:127.0.0.1/128", // IPv4-mapped address
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
}

func parseWeirdIPv4(s string) net.IP {
	parts := strings.Split(s, ".")
	if len(parts) > 4 {
		return nil
	}

	var nums []uint64
	for _, p := range parts {
		var n uint64
		var err error
		if strings.HasPrefix(p, "0x") || strings.HasPrefix(p, "0X") {
			n, err = strconv.ParseUint(p[2:], 16, 32)
		} else if strings.HasPrefix(p, "0") && len(p) > 1 {
			n, err = strconv.ParseUint(p[1:], 8, 32)
		} else {
			n, err = strconv.ParseUint(p, 10, 32)
		}
		if err != nil || n > 0xFFFFFFFF {
			return nil
		}
		nums = append(nums, n)
	}

	ip := uint32(0)
	switch len(nums) {
	case 1:
		ip = uint32(nums[0])
	case 2:
		ip = (uint32(nums[0]&0xFF) << 24) | uint32(nums[1])
	case 3:
		ip = (uint32(nums[0]&0xFF) << 24) | (uint32(nums[1]&0xFF) << 16) | uint32(nums[2])
	case 4:
		ip = (uint32(nums[0]&0xFF) << 24) | (uint32(nums[1]&0xFF) << 16) |
			(uint32(nums[2]&0xFF) << 8) | uint32(nums[3]&0xFF)
	default:
		return nil
	}

	b := []byte{byte(ip >> 24), byte(ip >> 16), byte(ip >> 8), byte(ip)}
	return net.IPv4(b[0], b[1], b[2], b[3])
}

// isPrivateIP checks if an IP address is within a private range.
func isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		parsedIP = parseWeirdIPv4(ip)
	}

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
