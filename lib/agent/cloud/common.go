package cloud

import (
	"encoding/json"
	. "main/aikido_types"
	"main/globals"
	. "main/globals"
	"main/log"
	"main/utils"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

func GetAgentInfo() AgentInfo {
	return AgentInfo{
		DryMode:   !utils.IsBlockingEnabled(),
		Hostname:  Machine.HostName,
		Version:   Version,
		IPAddress: Machine.IPAddress,
		OS: OsInfo{
			Name:    Machine.OS,
			Version: Machine.OSVersion,
		},
		Platform: PlatformInfo{
			Name:    EnvironmentConfig.PlatformName,
			Version: EnvironmentConfig.PlatformVersion,
		},
		Packages: make(map[string]string, 0),
		NodeEnv:  "",
		Library:  "firewall-php",
	}
}

func ResetHeartbeatTicker() {
	if !globals.CloudConfig.ReceivedAnyStats {
		log.Info("Resetting HeartBeatTicker to 1m!")
		HeartBeatTicker.Reset(1 * time.Minute)
	} else {
		if globals.CloudConfig.HeartbeatIntervalInMS >= globals.MinHeartbeatIntervalInMS {
			log.Infof("Resetting HeartBeatTicker to %dms!", globals.CloudConfig.HeartbeatIntervalInMS)
			HeartBeatTicker.Reset(time.Duration(globals.CloudConfig.HeartbeatIntervalInMS) * time.Millisecond)
		}
	}
}
func isWildcardEndpoint(route string) bool {
	return strings.Contains(route, "*")
}

func UpdateRateLimitingConfig() {
	globals.RateLimitingMutex.Lock()
	defer globals.RateLimitingMutex.Unlock()

	UpdatedEndpoints := map[RateLimitingKey]bool{}

	for _, newEndpointConfig := range globals.CloudConfig.Endpoints {
		k := RateLimitingKey{Method: newEndpointConfig.Method, Route: newEndpointConfig.Route}
		UpdatedEndpoints[k] = true

		rateLimitingData, exists := globals.RateLimitingMap[k]
		if exists {
			if rateLimitingData.Config.MaxRequests == newEndpointConfig.RateLimiting.MaxRequests &&
				rateLimitingData.Config.WindowSizeInMinutes == newEndpointConfig.RateLimiting.WindowSizeInMS*MinRateLimitingIntervalInMs {
				log.Debugf("New rate limiting endpoint config is the same: %v", newEndpointConfig)
				continue
			}

			log.Infof("Rate limiting endpoint config has changed: %v", newEndpointConfig)
			delete(globals.RateLimitingMap, k)
			delete(globals.RateLimitingWildcardMap, k)
		}

		if !newEndpointConfig.RateLimiting.Enabled {
			log.Infof("Got new rate limiting endpoint config, but is disabled: %v", newEndpointConfig)
			continue
		}

		if newEndpointConfig.RateLimiting.WindowSizeInMS < MinRateLimitingIntervalInMs ||
			newEndpointConfig.RateLimiting.WindowSizeInMS > MaxRateLimitingIntervalInMs {
			log.Warnf("Got new rate limiting endpoint config, but WindowSizeInMS is invalid: %v", newEndpointConfig)
			continue
		}

		log.Infof("Got new rate limiting endpoint config and storing to map: %v", newEndpointConfig)
		rateLimitingValue := &RateLimitingValue{
			Config: RateLimitingConfig{
				MaxRequests:         newEndpointConfig.RateLimiting.MaxRequests,
				WindowSizeInMinutes: newEndpointConfig.RateLimiting.WindowSizeInMS / MinRateLimitingIntervalInMs},
			UserCounts: make(map[string]*RateLimitingCounts),
			IpCounts:   make(map[string]*RateLimitingCounts),
		}

		if isWildcardEndpoint(k.Route) {
			routeRegex, err := regexp.Compile(strings.ReplaceAll(k.Route, "*", "(.*)") + "/?")
			if err != nil {
				log.Warnf("Route regex is not compiling: %s", k.Route)
			} else {
				log.Infof("Stored wildcard rate limiting config for: %v", k)
				globals.RateLimitingWildcardMap[k] = &RateLimitingWildcardValue{RouteRegex: routeRegex, RateLimitingValue: rateLimitingValue}
			}
		}
		log.Infof("Stored normal rate limiting config for: %v", k)
		globals.RateLimitingMap[k] = rateLimitingValue
	}

	for k := range globals.RateLimitingMap {
		_, exists := UpdatedEndpoints[k]
		if !exists {
			log.Infof("Removed rate limiting entry as it is no longer part of the config: %v", k)
			delete(globals.RateLimitingMap, k)
			delete(globals.RateLimitingWildcardMap, k)
		}
	}
}

func ApplyCloudConfig() {
	log.Infof("Applying new cloud config: %v", globals.CloudConfig)
	ResetHeartbeatTicker()
	UpdateRateLimitingConfig()
}

func UpdateIpsLists(BlockedIps []BlockedIpsData) map[string]IpBlocklist {
	m := make(map[string]IpBlocklist)
	for _, blockedIpsGroup := range BlockedIps {
		m[blockedIpsGroup.Source] = IpBlocklist{Description: blockedIpsGroup.Description, Ips: blockedIpsGroup.Ips}
	}
	return m
}

func UpdateListsConfig() bool {
	response, err := SendCloudRequest(globals.EnvironmentConfig.Endpoint, globals.ListsAPI, globals.ListsAPIMethod, nil)
	if err != nil {
		LogCloudRequestError("Error in sending lists request: ", err)
		return false
	}

	tempListsConfig := ListsConfigData{}
	err = json.Unmarshal(response, &tempListsConfig)
	if err != nil {
		log.Warnf("Failed to unmarshal lists config: %v", err)
		return false
	}

	CloudConfig.BlockedIpsList = UpdateIpsLists(tempListsConfig.BlockedIpAddresses)
	CloudConfig.MonitoredIpsList = UpdateIpsLists(tempListsConfig.MonitoredIpAddresses)
	CloudConfig.AllowedIpsList = UpdateIpsLists(tempListsConfig.AllowedIpAddresses)

	CloudConfig.BlockedUserAgents = tempListsConfig.BlockedUserAgents
	CloudConfig.MonitoredUserAgents = tempListsConfig.MonitoredUserAgents

	CloudConfig.UserAgentDetails = make(map[string]string)
	for _, userAgentDetail := range tempListsConfig.UserAgentDetails {
		CloudConfig.UserAgentDetails[userAgentDetail.Key] = userAgentDetail.Pattern
	}

	return true
}

func StoreCloudConfig(configReponse []byte) bool {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	tempCloudConfig := CloudConfigData{}
	err := json.Unmarshal(configReponse, &tempCloudConfig)
	if err != nil {
		log.Warnf("Failed to unmarshal cloud config!")
		return false
	}
	if tempCloudConfig.ConfigUpdatedAt <= globals.CloudConfig.ConfigUpdatedAt {
		log.Debugf("ConfigUpdatedAt is the same!")
		return true
	}
	globals.CloudConfig = tempCloudConfig
	UpdateListsConfig()
	ApplyCloudConfig()
	return true
}

func LogCloudRequestError(text string, err error) {
	if atomic.LoadUint32(&globals.GotTraffic) == 0 {
		// Wait for at least one request before we start logging any cloud request errors, including "no token set"
		// We need to do that because the token can be passed later via gRPC and the first request.
		return
	}
	if err.Error() == "no token set" {
		if atomic.LoadUint32(&globals.LoggedTokenError) != 0 {
			// Only report the "no token set" once, so we don't pollute the logs
			return
		}
		atomic.StoreUint32(&globals.LoggedTokenError, 1)
	}
	log.Warn(text, err)
}
