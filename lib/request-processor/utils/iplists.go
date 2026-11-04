package utils

import (
	. "main/aikido_types"
	"main/instance"
	"main/log"
	"net/netip"

	"go4.org/netipx"
)

type IpListMatch struct {
	Key         string
	Description string
}

func BuildIpList(description string, ipsList []string) (*IpList, error) {
	ipSet, err := BuildIpSet(ipsList)
	if err != nil {
		return nil, err
	}

	return &IpList{
		Description: description,
		IpSet:       *ipSet,
	}, nil
}

func BuildIpSet(ipsList []string) (*netipx.IPSet, error) {
	ipSetBuilder := netipx.IPSetBuilder{}

	for _, ip := range ipsList {
		prefix, err := netip.ParsePrefix(ip)
		if err == nil {
			ipSetBuilder.AddPrefix(prefix)
		} else {
			parsedIP, err := netip.ParseAddr(ip)
			if err == nil {
				ipSetBuilder.Add(parsedIP)
			} else {
				continue // Skip invalid IPs
			}
		}
	}

	ipSet, err := ipSetBuilder.IPSet()
	if err != nil {
		return nil, err
	}

	return ipSet, nil
}

func IsIpInSet(instance *instance.RequestProcessorInstance, ipSet *netipx.IPSet, ip string) int {
	if ipSet == nil || ipSet.Equal(&netipx.IPSet{}) {
		// No IPs configured in the list -> return default value
		return NoConfig
	}

	ipAddress, err := netip.ParseAddr(ip)
	if err != nil {
		log.Infof(instance, "Invalid ip address: %s\n", ip)
		return NoConfig
	}

	if ipSetContains(ipSet, ipAddress) {
		return Found
	}

	return NotFound
}

func IsIpInList(instance *instance.RequestProcessorInstance, ipList map[string]IpList, ip string) (int, []IpListMatch) {
	if len(ipList) == 0 {
		return NoConfig, []IpListMatch{}
	}

	ipAddress, err := netip.ParseAddr(ip)
	if err != nil {
		return NoConfig, []IpListMatch{}
	}
	matches := []IpListMatch{}
	for listKey, list := range ipList {
		if ipSetContains(&list.IpSet, ipAddress) {
			matches = append(matches, IpListMatch{Key: listKey, Description: list.Description})
		}
	}

	if len(matches) == 0 {
		return NotFound, matches
	}

	return Found, matches
}

// Requests and configured lists can represent the same IP address differently.
// Match IPv4, IPv4-mapped IPv6 and native IPv6 forms.
func ipSetContains(ipSet *netipx.IPSet, ipAddress netip.Addr) bool {
	if ipSet == nil {
		return false
	}

	ipAddress = ipAddress.Unmap()
	if ipSet.Contains(ipAddress) {
		return true
	}

	if ipAddress.Is4() {
		return ipSet.Contains(netip.AddrFrom16(ipAddress.As16()))
	}
	return false
}
