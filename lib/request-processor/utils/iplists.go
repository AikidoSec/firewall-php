package utils

import (
	"main/aikido_types"
	"net/netip"

	"go4.org/netipx"
)

func BuildIpList(description string, ipsList []string) (*aikido_types.IpList, error) {
	ipSet, err := BuildIpSet(ipsList)
	if err != nil {
		return nil, err
	}

	return &aikido_types.IpList{
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
