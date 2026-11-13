package utils

import (
	. "main/aikido_types"
	"math"
	"strconv"
	"strings"
)

func ParseTrustProxy(value string) TrustProxyConfig {
	if value == "" {
		return TrustProxyConfig{Enabled: true}
	}

	if strings.IndexFunc(value, func(c rune) bool { return c < '0' || c > '9' }) == -1 {
		count, err := strconv.Atoi(value)
		if err != nil {
			count = math.MaxInt
		}
		if count > 0 {
			return TrustProxyConfig{Enabled: true, Count: count}
		}
	}

	switch strings.ToLower(value) {
	case "true", "1", "yes", "y", "on":
		return TrustProxyConfig{Enabled: true}
	}

	if strings.ContainsAny(value, ".:") {
		ranges := strings.Split(value, ",")
		for i := range ranges {
			ranges[i] = strings.TrimSpace(ranges[i])
		}
		ipSet, err := BuildIpSet(ranges)
		if err == nil {
			return TrustProxyConfig{Enabled: true, Ranges: ipSet}
		}
	}
	return TrustProxyConfig{}
}
