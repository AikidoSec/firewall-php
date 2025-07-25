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

func buildIpList(cloudIpList map[string]*protos.IpList) map[string]IpList {
	ipList := map[string]IpList{}
	for ipListKey, protoIpList := range cloudIpList {
		ipSet, err := utils.BuildIpList(protoIpList.Description, protoIpList.Ips)
		if err != nil {
			log.Errorf("Error building IP list: %s\n", err)
			continue
		}
		ipList[ipListKey] = *ipSet
	}
	return ipList
}

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

func buildUserAgentsRegexpFromProto(userAgents string) *regexp.Regexp {
	if userAgents == "" {
		return nil
	}
	userAgentsRegexp, err := regexp.Compile("(?i)" + userAgents)
	if err != nil {
		log.Errorf("Error compiling user agents regex: %s\n", err)
		return nil
	}
	return userAgentsRegexp
}

func buildUserAgentDetailsFromProto(userAgentDetails map[string]string) map[string]*regexp.Regexp {
	m := map[string]*regexp.Regexp{}
	for key, value := range userAgentDetails {
		m[key] = buildUserAgentsRegexpFromProto(value)
	}
	return m
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

	globals.CloudConfig.BlockedIps = buildIpList(cloudConfigFromAgent.BlockedIps)
	globals.CloudConfig.MonitoredIps = buildIpList(cloudConfigFromAgent.MonitoredIps)

	globals.CloudConfig.BlockedUserAgents = buildUserAgentsRegexpFromProto(cloudConfigFromAgent.BlockedUserAgents)
	globals.CloudConfig.MonitoredUserAgents = buildUserAgentsRegexpFromProto(cloudConfigFromAgent.MonitoredUserAgents)

	globals.CloudConfig.AllowedIps = buildIpList(cloudConfigFromAgent.AllowedIps)

	globals.CloudConfig.UserAgentDetails = buildUserAgentDetailsFromProto(cloudConfigFromAgent.UserAgentDetails)

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
