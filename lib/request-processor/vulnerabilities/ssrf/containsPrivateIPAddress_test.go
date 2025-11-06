package ssrf

import (
	"strings"
	"testing"
)

var publicIPs = []string{
	"44.37.112.180",
	"46.192.247.73",
	"71.12.102.112",
	"101.0.26.90",
	"111.211.73.40",
	"156.238.194.84",
	"164.101.185.82",
	"223.231.138.242",
	"::1fff:0.0.0.0",
	"::1fff:10.0.0.0",
	"::1fff:0:0.0.0.0",
	"::1fff:0:10.0.0.0",
	"2001:2:ffff:ffff:ffff:ffff:ffff:ffff",
	"64:ff9a::0.0.0.0",
	"64:ff9a::255.255.255.255",
	"99::",
	"99::ffff:ffff:ffff:ffff",
	"101::",
	"101::ffff:ffff:ffff:ffff",
	"2000::",
	"2000::ffff:ffff:ffff:ffff:ffff:ffff",
	"2001:10::",
	"2001:1f:ffff:ffff:ffff:ffff:ffff:ffff",
	"2001:db7::",
	"2001:db7:ffff:ffff:ffff:ffff:ffff:ffff",
	"2001:db9::",
	"fb00::",
	"fbff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
	"fec0::",
}

var privateIPs = []string{
	"localhost",
	"0.0.0.0",
	//"0000.0000.0000.0000",
	//"0000.0000",
	"0.0.0.1",
	"0.0.0.7",
	"0.0.0.255",
	"0.0.255.255",
	"0.1.255.255",
	"0.15.255.255",
	"0.63.255.255",
	"0.255.255.254",
	"0.255.255.255",
	"10.0.0.0",
	"10.0.0.1",
	//"10.0.0.01",
	//"10.0.0.001",
	"10.255.255.254",
	"10.255.255.255",
	"100.64.0.0",
	"100.64.0.1",
	"100.127.255.254",
	"100.127.255.255",
	"127.0.0.0",
	"127.0.0.1",
	//"//127.0.0.01",
	//"127.1",
	//"127.0.1",
	//"127.000.000.1",
	"127.255.255.254",
	"127.255.255.255",
	"169.254.0.0",
	"169.254.0.1",
	"169.254.255.254",
	"169.254.255.255",
	"172.16.0.0",
	"172.16.0.1",
	//"172.16.0.001",
	"172.31.255.254",
	"172.31.255.255",
	"192.0.0.0",
	"192.0.0.1",
	"192.0.0.6",
	"192.0.0.7",
	"192.0.0.8",
	"192.0.0.9",
	"192.0.0.10",
	"192.0.0.11",
	"192.0.0.170",
	"192.0.0.171",
	"192.0.0.254",
	"192.0.0.255",
	"192.0.2.0",
	"192.0.2.1",
	"192.0.2.254",
	"192.0.2.255",
	"192.31.196.0",
	"192.31.196.1",
	"192.31.196.254",
	"192.31.196.255",
	"192.52.193.0",
	"192.52.193.1",
	"192.52.193.254",
	"192.52.193.255",
	"192.88.99.0",
	"192.88.99.1",
	"192.88.99.254",
	"192.88.99.255",
	"192.168.0.0",
	"192.168.0.1",
	"192.168.255.254",
	"192.168.255.255",
	"192.175.48.0",
	"192.175.48.1",
	"192.175.48.254",
	"192.175.48.255",
	"198.18.0.0",
	"198.18.0.1",
	"198.19.255.254",
	"198.19.255.255",
	"198.51.100.0",
	"198.51.100.1",
	"198.51.100.254",
	"198.51.100.255",
	"203.0.113.0",
	"203.0.113.1",
	"203.0.113.254",
	"203.0.113.255",
	"240.0.0.0",
	"240.0.0.1",
	"224.0.0.0",
	"224.0.0.1",
	"255.0.0.0",
	"255.192.0.0",
	"255.240.0.0",
	"255.254.0.0",
	"255.255.0.0",
	"255.255.255.0",
	"255.255.255.248",
	"255.255.255.254",
	"255.255.255.255",
	"0000:0000:0000:0000:0000:0000:0000:0000",
	"::",
	"::1",
	"::ffff:0.0.0.0",
	"::ffff:127.0.0.1",
	"fe80::",
	"fe80::1",
	"fe80::abc:1",
	"febf:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
	"fc00::",
	"fc00::1",
	"fc00::abc:1",
	"fdff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
	//"2130706433",
	//"0x7f000001",

	// AWS metadata
	"fd00:ec2::254",
	"169.254.169.254",
}

var invalidIPs = []string{
	"100::ffff::",
	"::ffff:0.0.255.255.255",
	"::ffff:0.255.255.255.255",
}

// TestPublicIPs checks that public IPs return false for private IP check.
func TestPublicIPs(t *testing.T) {
	for _, ip := range publicIPs {
		if strings.Contains(ip, ":") {
			ip = "[" + ip + "]" // Enclose IPv6 addresses in brackets
		}
		if containsPrivateIPAddress(ip) {
			t.Errorf("Expected %s to be public, but was detected as private", ip)
		}
	}
}

// TestPrivateIPs checks that private IPs return true for private IP check.
/*
func TestPrivateIPs(t *testing.T) {
	for _, ip := range privateIPs {
		if strings.Contains(ip, ":") {
			ip = "[" + ip + "]" // Enclose IPv6 addresses in brackets
		}
		if !containsPrivateIPAddress(ip) {
			t.Errorf("Expected %s to be private, but was detected as public", ip)
		}
	}
}
*/

// TestInvalidIPs checks that invalid IPs return false for private IP check.
func TestInvalidIPs(t *testing.T) {
	for _, ip := range invalidIPs {
		if strings.Contains(ip, ":") {
			ip = "[" + ip + "]" // Enclose IPv6 addresses in brackets
		}
		if containsPrivateIPAddress(ip) {
			t.Errorf("Expected %s to be invalid, but was detected as private/public", ip)
		}
	}
}
