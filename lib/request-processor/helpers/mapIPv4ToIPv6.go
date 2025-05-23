package helpers

import (
	"fmt"
	"strconv"
	"strings"
)

// Maps an IPv4 address to an IPv6 address.
// e.g. 127.0.0.0/8 -> ::ffff:127.0.0.0/104
func MapIPv4ToIPv6(ip string) string {
	if !strings.Contains(ip, "/") {
		// No CIDR suffix, assume /32
		return fmt.Sprintf("::ffff:%s/128", ip)
	}

	parts := strings.Split(ip, "/")
	suffix, err := strconv.Atoi(parts[1])
	if err != nil {
		// Throw an error if the suffix is not a number
		// Because the input is statically defined in the code, this should never happen to a end user
		panic(fmt.Sprintf("Invalid CIDR suffix: %s", parts[1]))
	}

	// we add 96 to the suffix, since ::ffff: already is 96 bits, so the 32 remaining bits are decided by the IPv4 address
	return fmt.Sprintf("::ffff:%s/%d", parts[0], suffix+96)
}
