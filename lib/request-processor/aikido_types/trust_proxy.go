package aikido_types

import (
	"net/netip"

	"go4.org/netipx"
)

type TrustProxyConfig struct {
	Enabled bool
	Count   int
	Ranges  *netipx.IPSet
}

func (config TrustProxyConfig) Contains(ip string) bool {
	if config.Ranges == nil {
		return false
	}
	address, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}
	return config.Ranges.Contains(address) || config.Ranges.Contains(address.Unmap())
}

func (config TrustProxyConfig) IsTrusted(remoteAddress string) bool {
	if !config.Enabled {
		return false
	}
	if config.Ranges != nil {
		return config.Contains(remoteAddress)
	}
	return true
}
