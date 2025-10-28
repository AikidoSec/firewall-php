package grpc

import (
	. "main/aikido_types"
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

func storeEndpointConfig(server *ServerData, ep *protos.Endpoint) {
	server.CloudConfig.Endpoints[EndpointKey{Method: ep.Method, Route: ep.Route}] = getEndpointData(ep)
}

func storeWildcardEndpointConfig(server *ServerData, ep *protos.Endpoint) {
	wildcardRouteCompiled, err := regexp.Compile(strings.ReplaceAll(ep.Route, "*", ".*"))
	if err != nil {
		return
	}

	wildcardRoutes, exists := server.CloudConfig.WildcardEndpoints[ep.Method]
	if !exists {
		server.CloudConfig.WildcardEndpoints[ep.Method] = []WildcardEndpointData{}
	}

	server.CloudConfig.WildcardEndpoints[ep.Method] = append(wildcardRoutes, WildcardEndpointData{RouteRegex: wildcardRouteCompiled, Data: getEndpointData(ep)})
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

func setCloudConfig(server *ServerData, cloudConfigFromAgent *protos.CloudConfig) {
	if cloudConfigFromAgent == nil {
		return
	}

	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()

	server.CloudConfig.ConfigUpdatedAt = cloudConfigFromAgent.ConfigUpdatedAt

	server.CloudConfig.Endpoints = map[EndpointKey]EndpointData{}
	server.CloudConfig.WildcardEndpoints = map[string][]WildcardEndpointData{}

	for _, ep := range cloudConfigFromAgent.Endpoints {
		if utils.IsWildcardEndpoint(ep.Method, ep.Route) {
			storeWildcardEndpointConfig(server, ep)
		} else {
			storeEndpointConfig(server, ep)
		}
	}

	server.CloudConfig.BlockedUserIds = map[string]bool{}
	for _, userId := range cloudConfigFromAgent.BlockedUserIds {
		server.CloudConfig.BlockedUserIds[userId] = true
	}

	bypassedIPSet, bypassedIPSetErr := utils.BuildIpSet(cloudConfigFromAgent.BypassedIps)
	server.CloudConfig.BypassedIps = bypassedIPSet
	if bypassedIPSet == nil {
		log.Errorf("Error building bypassed IP set: %s\n", bypassedIPSetErr)
	}

	if cloudConfigFromAgent.Block {
		server.CloudConfig.Block = 1
	} else {
		server.CloudConfig.Block = 0
	}

	server.CloudConfig.BlockedIps = buildIpList(cloudConfigFromAgent.BlockedIps)
	server.CloudConfig.MonitoredIps = buildIpList(cloudConfigFromAgent.MonitoredIps)

	server.CloudConfig.BlockedUserAgents = buildUserAgentsRegexpFromProto(cloudConfigFromAgent.BlockedUserAgents)
	server.CloudConfig.MonitoredUserAgents = buildUserAgentsRegexpFromProto(cloudConfigFromAgent.MonitoredUserAgents)

	server.CloudConfig.AllowedIps = buildIpList(cloudConfigFromAgent.AllowedIps)

	server.CloudConfig.UserAgentDetails = buildUserAgentDetailsFromProto(cloudConfigFromAgent.UserAgentDetails)

	// Force garbage collection to ensure that the IP blocklists temporary memory is released ASAP
	runtime.GC()
}

func StartCloudConfigRoutine() {
	GetCloudConfigForAllServers(5 * time.Second)

	stopChan = make(chan struct{})

	go func() {
		for {
			select {
			case <-cloudConfigTicker.C:
				GetCloudConfigForAllServers(5 * time.Second)
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
