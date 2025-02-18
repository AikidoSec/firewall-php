package grpc

import (
	. "main/aikido_types"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"regexp"
	"strings"
	"time"

	"github.com/seancfoley/ipaddress-go/ipaddr"
)

var (
	stopChan          chan struct{}
	cloudConfigTicker = time.NewTicker(1 * time.Minute)
)

func buildIpBlocklist(name, description string, ipsList []string) IpBlockList {
	ipBlocklist := IpBlockList{
		Description: description,
		TrieV4:      &ipaddr.IPv4AddressTrie{},
		TrieV6:      &ipaddr.IPv6AddressTrie{},
	}

	for _, ip := range ipsList {
		ipAddress, err := ipaddr.NewIPAddressString(ip).ToAddress()
		if err != nil {
			log.Infof("Invalid address for %s: %s\n", name, ip)
			continue
		}

		if ipAddress.IsIPv4() {
			ipBlocklist.TrieV4.Add(ipAddress.ToIPv4())
		} else if ipAddress.IsIPv6() {
			ipBlocklist.TrieV6.Add(ipAddress.ToIPv6())
		}
	}

	log.Debugf("%s (v4): %v", name, ipBlocklist.TrieV4)
	log.Debugf("%s (v6): %v", name, ipBlocklist.TrieV6)
	return ipBlocklist
}

func getEndpointData(ep *protos.Endpoint) EndpointData {
	endpointData := EndpointData{
		ForceProtectionOff: ep.ForceProtectionOff,
		RateLimiting: RateLimiting{
			Enabled: ep.RateLimiting.Enabled,
		},
		AllowedIPAddresses: map[string]bool{},
	}
	for _, ip := range ep.AllowedIPAddresses {
		endpointData.AllowedIPAddresses[ip] = true
	}
	return endpointData
}

func storeEndpointConfig(ep *protos.Endpoint) {
	globals.CloudConfig.Endpoints[EndpointKey{Method: ep.Method, Route: ep.Route}] = getEndpointData(ep)
}

func storeWildcardEndpointConfig(ep *protos.Endpoint) {
	wildcardRouteCompiled, err := regexp.Compile(strings.ReplaceAll(ep.Route, "*", ".*"))
	if err != nil {
		return
	}

	wildcardRoutes, exists := globals.CloudConfig.WildcardEndpoints[ep.Method]
	if !exists {
		globals.CloudConfig.WildcardEndpoints[ep.Method] = []WildcardEndpointData{}
	}

	globals.CloudConfig.WildcardEndpoints[ep.Method] = append(wildcardRoutes, WildcardEndpointData{RouteRegex: wildcardRouteCompiled, Data: getEndpointData(ep)})
}

func isWildcardEndpoint(method, route string) bool {
	return method == "*" || strings.Contains(route, "*")
}

func setCloudConfig(cloudConfigFromAgent *protos.CloudConfig) {
	if cloudConfigFromAgent == nil {
		return
	}

	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	globals.CloudConfig.ConfigUpdatedAt = cloudConfigFromAgent.ConfigUpdatedAt

	globals.CloudConfig.Endpoints = map[EndpointKey]EndpointData{}
	globals.CloudConfig.WildcardEndpoints = map[string][]WildcardEndpointData{}

	for _, ep := range cloudConfigFromAgent.Endpoints {
		if isWildcardEndpoint(ep.Method, ep.Route) {
			storeWildcardEndpointConfig(ep)
		} else {
			storeEndpointConfig(ep)
		}
	}

	globals.CloudConfig.BlockedUserIds = map[string]bool{}
	for _, userId := range cloudConfigFromAgent.BlockedUserIds {
		globals.CloudConfig.BlockedUserIds[userId] = true
	}

	globals.CloudConfig.BypassedIps = map[string]bool{}
	for _, ip := range cloudConfigFromAgent.BypassedIps {
		globals.CloudConfig.BypassedIps[ip] = true
	}

	if cloudConfigFromAgent.Block {
		globals.CloudConfig.Block = 1
	} else {
		globals.CloudConfig.Block = 0
	}

	globals.CloudConfig.BlockedIps = map[string]IpBlockList{}
	for ipBlocklistSource, ipBlocklist := range cloudConfigFromAgent.BlockedIps {
		globals.CloudConfig.BlockedIps[ipBlocklistSource] = buildIpBlocklist(ipBlocklistSource, ipBlocklist.Description, ipBlocklist.Ips)
	}

	if cloudConfigFromAgent.BlockedUserAgents != "" {
		globals.CloudConfig.BlockedUserAgents, _ = regexp.Compile("(?i)" + cloudConfigFromAgent.BlockedUserAgents)
	} else {
		globals.CloudConfig.BlockedUserAgents = nil
	}

}

func startCloudConfigRoutine() {
	GetCloudConfig()

	stopChan = make(chan struct{})

	go func() {
		for {
			select {
			case <-cloudConfigTicker.C:
				GetCloudConfig()
			case <-stopChan:
				cloudConfigTicker.Stop()
				return
			}
		}
	}()
}

func stopCloudConfigRoutine() {
	if stopChan != nil {
		close(stopChan)
	}
}
