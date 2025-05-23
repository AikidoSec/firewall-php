package grpc

import (
	. "main/aikido_types"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"main/utils"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var (
	stopChan          chan struct{}
	cloudConfigTicker = time.NewTicker(1 * time.Minute)
)

func getEndpointData(ep *protos.Endpoint) EndpointData {
	allowedIPSet, err := utils.BuildIpSet(ep.AllowedIPAddresses)
	if err != nil {
		log.Errorf("Error building allowed IP set: %s\n", err)
	}
	endpointData := EndpointData{
		ForceProtectionOff: ep.ForceProtectionOff,
		RateLimiting: RateLimiting{
			Enabled: ep.RateLimiting.Enabled,
		},
		AllowedIPAddresses: allowedIPSet,
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
		if utils.IsWildcardEndpoint(ep.Method, ep.Route) {
			storeWildcardEndpointConfig(ep)
		} else {
			storeEndpointConfig(ep)
		}
	}

	globals.CloudConfig.BlockedUserIds = map[string]bool{}
	for _, userId := range cloudConfigFromAgent.BlockedUserIds {
		globals.CloudConfig.BlockedUserIds[userId] = true
	}

	bypassedIPSet, bypassedIPSetErr := utils.BuildIpSet(cloudConfigFromAgent.BypassedIps)
	globals.CloudConfig.BypassedIps = bypassedIPSet
	if bypassedIPSet == nil {
		log.Errorf("Error building bypassed IP set: %s\n", bypassedIPSetErr)
	}

	if cloudConfigFromAgent.Block {
		globals.CloudConfig.Block = 1
	} else {
		globals.CloudConfig.Block = 0
	}

	globals.CloudConfig.BlockedIps = map[string]IpBlockList{}
	for ipBlocklistSource, ipBlocklist := range cloudConfigFromAgent.BlockedIps {
		ipBlocklist, err := utils.BuildIpBlocklist(ipBlocklistSource, ipBlocklist.Description, ipBlocklist.Ips)
		if err != nil {
			log.Errorf("Error building IP blocklist: %s\n", err)
			continue
		}
		globals.CloudConfig.BlockedIps[ipBlocklistSource] = *ipBlocklist
	}

	if cloudConfigFromAgent.BlockedUserAgents != "" {
		globals.CloudConfig.BlockedUserAgents, _ = regexp.Compile("(?i)" + cloudConfigFromAgent.BlockedUserAgents)
	} else {
		globals.CloudConfig.BlockedUserAgents = nil
	}

	// Force garbage collection to ensure that the IP blocklists temporary memory is released ASAP
	runtime.GC()
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
