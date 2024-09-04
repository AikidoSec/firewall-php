package cloud

import (
	"encoding/json"
	. "main/aikido_types"
	"main/globals"
	. "main/globals"
	"main/log"
	"main/utils"
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
		globals.RateLimitingMap[k] = &RateLimitingValue{
			Config: RateLimitingConfig{
				MaxRequests:         newEndpointConfig.RateLimiting.MaxRequests,
				WindowSizeInMinutes: newEndpointConfig.RateLimiting.WindowSizeInMS / MinRateLimitingIntervalInMs},
		}
	}

	for k := range globals.RateLimitingMap {
		_, exists := UpdatedEndpoints[k]
		if !exists {
			log.Infof("Removed rate limiting entry as it is no longer part of the config: %v", k)
			delete(globals.RateLimitingMap, k)
		}
	}
}

func ApplyCloudConfig() {
	log.Infof("Applying new cloud config: %v", globals.CloudConfig)
	ResetHeartbeatTicker()
	UpdateRateLimitingConfig()
}

func StoreCloudConfig(response []byte) bool {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	tempCloudConfig := CloudConfigData{}
	err := json.Unmarshal(response, &tempCloudConfig)
	if err != nil {
		log.Warnf("Failed to unmarshal cloud config!")
		return false
	}
	if tempCloudConfig.ConfigUpdatedAt <= globals.CloudConfig.ConfigUpdatedAt {
		log.Debugf("ConfigUpdatedAt is the same!")
		return true
	}
	globals.CloudConfig = tempCloudConfig
	ApplyCloudConfig()
	return true
}
