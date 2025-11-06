package cloud

import (
	"encoding/json"
	. "main/aikido_types"
	"main/constants"
	"main/globals"
	"main/log"
	"main/utils"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

func GetAgentInfo(server *ServerData) AgentInfo {
	return AgentInfo{
		DryMode:   !utils.IsBlockingEnabled(server),
		Hostname:  globals.Machine.HostName,
		Version:   constants.Version,
		IPAddress: globals.Machine.IPAddress,
		OS: OsInfo{
			Name:    globals.Machine.OS,
			Version: globals.Machine.OSVersion,
		},
		Platform: PlatformInfo{
			Name:    server.AikidoConfig.PlatformName,
			Version: server.AikidoConfig.PlatformVersion,
		},
		Packages: make(map[string]string, 0),
		NodeEnv:  "",
		Library:  "firewall-php",
	}
}

func ResetHeartbeatTicker(server *ServerData) {
	if !server.CloudConfig.ReceivedAnyStats {
		log.Info(server.Logger, "Resetting HeartbeatTicker to 1m!")
		server.PollingData.HeartbeatTicker.Reset(1 * time.Minute)
	} else {
		if server.CloudConfig.HeartbeatIntervalInMS >= constants.MinHeartbeatIntervalInMS {
			log.Infof(server.Logger, "Resetting HeartbeatTicker to %dms!", server.CloudConfig.HeartbeatIntervalInMS)
			server.PollingData.HeartbeatTicker.Reset(time.Duration(server.CloudConfig.HeartbeatIntervalInMS) * time.Millisecond)
		}
	}
}

func isWildcardEndpoint(route string) bool {
	return strings.Contains(route, "*")
}

func UpdateRateLimitingConfig(server *ServerData) {
	server.RateLimitingMutex.Lock()
	defer server.RateLimitingMutex.Unlock()

	UpdatedEndpoints := map[RateLimitingKey]bool{}

	for _, newEndpointConfig := range server.CloudConfig.Endpoints {
		k := RateLimitingKey{Method: newEndpointConfig.Method, Route: newEndpointConfig.Route}
		UpdatedEndpoints[k] = true

		rateLimitingData, exists := server.RateLimitingMap[k]
		if exists {
			if rateLimitingData.Config.MaxRequests == newEndpointConfig.RateLimiting.MaxRequests &&
				rateLimitingData.Config.WindowSizeInMinutes == newEndpointConfig.RateLimiting.WindowSizeInMS*constants.MinRateLimitingIntervalInMs {
				log.Debugf(server.Logger, "New rate limiting endpoint config is the same: %v", newEndpointConfig)
				continue
			}

			log.Infof(server.Logger, "Rate limiting endpoint config has changed: %v", newEndpointConfig)
			delete(server.RateLimitingMap, k)
			delete(server.RateLimitingWildcardMap, k)
		}

		if !newEndpointConfig.RateLimiting.Enabled {
			log.Infof(server.Logger, "Got new rate limiting endpoint config, but is disabled: %v", newEndpointConfig)
			continue
		}

		if newEndpointConfig.RateLimiting.WindowSizeInMS < constants.MinRateLimitingIntervalInMs ||
			newEndpointConfig.RateLimiting.WindowSizeInMS > constants.MaxRateLimitingIntervalInMs {
			log.Warnf(server.Logger, "Got new rate limiting endpoint config, but WindowSizeInMS is invalid: %v", newEndpointConfig)
			continue
		}

		log.Infof(server.Logger, "Got new rate limiting endpoint config and storing to map: %v", newEndpointConfig)
		rateLimitingValue := &RateLimitingValue{
			Method: k.Method,
			Route:  k.Route,
			Config: RateLimitingConfig{
				MaxRequests:         newEndpointConfig.RateLimiting.MaxRequests,
				WindowSizeInMinutes: newEndpointConfig.RateLimiting.WindowSizeInMS / constants.MinRateLimitingIntervalInMs},
			UserCounts:           make(map[string]*SlidingWindow),
			IpCounts:             make(map[string]*SlidingWindow),
			RateLimitGroupCounts: make(map[string]*SlidingWindow),
		}

		if isWildcardEndpoint(k.Route) {
			routeRegex, err := regexp.Compile(strings.ReplaceAll(k.Route, "*", "(.*)") + "/?")
			if err != nil {
				log.Warnf(server.Logger, "Route regex is not compiling: %s", k.Route)
			} else {
				log.Infof(server.Logger, "Stored wildcard rate limiting config for: %v", k)
				server.RateLimitingWildcardMap[k] = &RateLimitingWildcardValue{RouteRegex: routeRegex, RateLimitingValue: rateLimitingValue}
			}
		}
		log.Infof(server.Logger, "Stored normal rate limiting config for: %v", k)
		server.RateLimitingMap[k] = rateLimitingValue
	}

	for k := range server.RateLimitingMap {
		_, exists := UpdatedEndpoints[k]
		if !exists {
			log.Infof(server.Logger, "Removed rate limiting entry as it is no longer part of the config: %v", k)
			delete(server.RateLimitingMap, k)
			delete(server.RateLimitingWildcardMap, k)
		}
	}
}

func ApplyCloudConfig(server *ServerData) {
	log.Infof(server.Logger, "Applying new cloud config: %v", server.CloudConfig)
	ResetHeartbeatTicker(server)
	UpdateRateLimitingConfig(server)
}

func UpdateIpsLists(ipLists []IpsData) map[string]IpBlocklist {
	m := make(map[string]IpBlocklist)
	for _, ipList := range ipLists {
		m[ipList.Key] = IpBlocklist{Description: ipList.Description, Ips: ipList.Ips}
	}
	return m
}

func UpdateListsConfig(server *ServerData) bool {
	response, err := SendCloudRequest(server, server.AikidoConfig.Endpoint, constants.ListsAPI, constants.ListsAPIMethod, nil)
	if err != nil {
		LogCloudRequestError(server, "Error in sending lists request: ", err)
		return false
	}

	tempListsConfig := ListsConfigData{}
	err = json.Unmarshal(response, &tempListsConfig)
	if err != nil {
		log.Warnf(server.Logger, "Failed to unmarshal lists config: %v", err)
		return false
	}

	server.CloudConfig.BlockedIpsList = UpdateIpsLists(tempListsConfig.BlockedIpAddresses)
	server.CloudConfig.MonitoredIpsList = UpdateIpsLists(tempListsConfig.MonitoredIpAddresses)
	server.CloudConfig.AllowedIpsList = UpdateIpsLists(tempListsConfig.AllowedIpAddresses)

	server.CloudConfig.BlockedUserAgents = tempListsConfig.BlockedUserAgents
	server.CloudConfig.MonitoredUserAgents = tempListsConfig.MonitoredUserAgents

	server.CloudConfig.UserAgentDetails = make(map[string]string)
	for _, userAgentDetail := range tempListsConfig.UserAgentDetails {
		server.CloudConfig.UserAgentDetails[userAgentDetail.Key] = userAgentDetail.Pattern
	}

	/* Force garbage collection to ensure that the IP blocklists temporary memory is released ASAP */
	tempListsConfig = ListsConfigData{}
	runtime.GC()

	return true
}

func StoreCloudConfig(server *ServerData, configReponse []byte) bool {
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()

	tempCloudConfig := CloudConfigData{}
	err := json.Unmarshal(configReponse, &tempCloudConfig)
	if err != nil {
		log.Warnf(server.Logger, "Failed to unmarshal cloud config: %v", err)
		return false
	}
	if tempCloudConfig.ConfigUpdatedAt <= server.CloudConfig.ConfigUpdatedAt {
		log.Debugf(server.Logger, "ConfigUpdatedAt is the same!")
		return true
	}
	server.CloudConfig = tempCloudConfig
	UpdateListsConfig(server)
	ApplyCloudConfig(server)
	return true
}

func LogCloudRequestError(server *ServerData, text string, err error) {
	if atomic.LoadUint32(&server.GotTraffic) == 0 {
		// Wait for at least one request before we start logging any cloud request errors, including "no token set"
		// We need to do that because the token can be passed later via gRPC and the first request.
		return
	}
	if err.Error() == "no token set" {
		if atomic.LoadUint32(&server.LoggedTokenError) != 0 {
			// Only report the "no token set" once, so we don't pollute the logs
			return
		}
		atomic.StoreUint32(&server.LoggedTokenError, 1)
	}
	log.Warn(server.Logger, text, err)
}
