package grpc

import (
	. "main/aikido_types"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"main/utils"
	"regexp"
	"runtime"
	"time"
)

var (
	stopChan          chan struct{}
	cloudConfigTicker = time.NewTicker(1 * time.Minute)
)

func buildIpListFromProto(monitoredIpsList map[string]*protos.IpBlockList) map[string]IpBlockList {
	m := map[string]IpBlockList{}
	for ipBlocklistSource, ipBlocklist := range monitoredIpsList {
		ipBlocklist, err := utils.BuildIpBlocklist(ipBlocklistSource, ipBlocklist.Description, ipBlocklist.Ips)
		if err != nil {
			log.Errorf("Error building IP blocklist: %s\n", err)
			continue
		}
		m[ipBlocklistSource] = *ipBlocklist
	}
	return m
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

func setCloudConfig(cloudConfigFromAgent *protos.CloudConfig) {
	if cloudConfigFromAgent == nil {
		return
	}

	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	globals.CloudConfig.ConfigUpdatedAt = cloudConfigFromAgent.ConfigUpdatedAt

	globals.CloudConfig.Endpoints = map[EndpointKey]EndpointData{}
	for _, ep := range cloudConfigFromAgent.Endpoints {

		allowedIPSet, allowedIPSetErr := utils.BuildIpSet(ep.AllowedIPAddresses)
		if allowedIPSet == nil {
			log.Errorf("Error building allowed IP set: %s\n", allowedIPSetErr)
		}

		endpointData := EndpointData{
			ForceProtectionOff: ep.ForceProtectionOff,
			RateLimiting: RateLimiting{
				Enabled: ep.RateLimiting.Enabled,
			},
			AllowedIPAddresses: allowedIPSet,
		}

		globals.CloudConfig.Endpoints[EndpointKey{Method: ep.Method, Route: ep.Route}] = endpointData
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

	globals.CloudConfig.BlockedIps = buildIpListFromProto(cloudConfigFromAgent.BlockedIps)
	globals.CloudConfig.MonitoredIps = buildIpListFromProto(cloudConfigFromAgent.MonitoredIps)

	globals.CloudConfig.BlockedUserAgents = buildUserAgentsRegexpFromProto(cloudConfigFromAgent.BlockedUserAgents)
	globals.CloudConfig.MonitoredUserAgents = buildUserAgentsRegexpFromProto(cloudConfigFromAgent.MonitoredUserAgents)

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
