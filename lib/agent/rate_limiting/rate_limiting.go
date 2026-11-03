package rate_limiting

import (
	. "main/aikido_types"
	"main/constants"
	"main/log"
	"main/utils"
	"regexp"
	"strings"
	"time"
)

// The caller must hold endpoint.Mutex to keep counters and reset time consistent.
func AdvanceEndpointQueues(endpoint *RateLimitingValue, now time.Time) {
	if now.Before(endpoint.NextResetAt) {
		return
	}
	elapsedMinutes := int(now.Sub(endpoint.NextResetAt)/time.Minute) + 1
	if elapsedMinutes >= endpoint.Config.WindowSizeInMinutes {
		clear(endpoint.UserCounts)
		clear(endpoint.IpCounts)
		clear(endpoint.RateLimitGroupCounts)
	} else {
		for range elapsedMinutes {
			AdvanceSlidingWindowMap(endpoint.UserCounts, endpoint.Config.WindowSizeInMinutes)
			AdvanceSlidingWindowMap(endpoint.IpCounts, endpoint.Config.WindowSizeInMinutes)
			AdvanceSlidingWindowMap(endpoint.RateLimitGroupCounts, endpoint.Config.WindowSizeInMinutes)
		}
	}
	endpoint.NextResetAt = endpoint.NextResetAt.Add(time.Duration(elapsedMinutes) * time.Minute)
}

func AdvanceRateLimitingQueues(server *ServerData) {
	server.RateLimitingMutex.RLock()
	// Config updates replace the whole map, so we can iterate this map after unlocking.
	endpoints := server.RateLimitingMap
	server.RateLimitingMutex.RUnlock()

	for _, endpoint := range endpoints {
		endpoint.Mutex.Lock()
		AdvanceEndpointQueues(endpoint, time.Now())
		endpoint.Mutex.Unlock()
	}
}

// The caller must hold server.CloudConfigMutex to prevent endpoint settings changing during the update.
func UpdateConfig(server *ServerData) {
	server.RateLimitingMutex.Lock()
	defer server.RateLimitingMutex.Unlock()

	newMap := make(map[RateLimitingKey]*RateLimitingValue)
	newWildcardMap := make(map[RateLimitingKey]*RateLimitingWildcardValue)

	for _, newEndpointConfig := range server.CloudConfig.Endpoints {
		// Skip disabled endpoints
		if !newEndpointConfig.RateLimiting.Enabled {
			log.Infof(server.Logger, "Got new rate limiting endpoint config, but is disabled: %v", newEndpointConfig)
			continue
		}

		// Validate window size
		if newEndpointConfig.RateLimiting.WindowSizeInMS < constants.MinRateLimitingIntervalInMs ||
			newEndpointConfig.RateLimiting.WindowSizeInMS > constants.MaxRateLimitingIntervalInMs {
			log.Warnf(server.Logger, "Got new rate limiting endpoint config, but WindowSizeInMS is invalid: %v", newEndpointConfig)
			continue
		}

		k := RateLimitingKey{Method: newEndpointConfig.Method, Route: newEndpointConfig.Route}
		newConfig := RateLimitingConfig{
			MaxRequests:         newEndpointConfig.RateLimiting.MaxRequests,
			WindowSizeInMinutes: newEndpointConfig.RateLimiting.WindowSizeInMS / constants.MinRateLimitingIntervalInMs,
		}

		// Check if we can preserve existing data
		if existingData, exists := server.RateLimitingMap[k]; exists {
			if existingData.Config.MaxRequests == newConfig.MaxRequests &&
				existingData.Config.WindowSizeInMinutes == newConfig.WindowSizeInMinutes {
				// Config unchanged, preserve data
				log.Debugf(server.Logger, "New rate limiting endpoint config is the same: %v", newEndpointConfig)
				newMap[k] = existingData
				// Also preserve wildcard if it exists
				if wildcardData, wildcardExists := server.RateLimitingWildcardMap[k]; wildcardExists {
					newWildcardMap[k] = wildcardData
				}
				continue
			}
			log.Infof(server.Logger, "Rate limiting endpoint config has changed: %v", newEndpointConfig)
		} else {
			log.Infof(server.Logger, "Got new rate limiting endpoint config and storing to map: %v", newEndpointConfig)
		}

		// Create new entry
		rateLimitingValue := &RateLimitingValue{
			Method:               k.Method,
			Route:                k.Route,
			Config:               newConfig,
			UserCounts:           make(map[string]*SlidingWindow),
			IpCounts:             make(map[string]*SlidingWindow),
			RateLimitGroupCounts: make(map[string]*SlidingWindow),
			NextResetAt:          time.Now().Add(time.Minute),
		}

		// Handle wildcard routes
		if isWildcardEndpoint(k.Route) {
			routeRegex, err := regexp.Compile(strings.ReplaceAll(k.Route, "*", "(.*)") + "/?")
			if err != nil {
				log.Warnf(server.Logger, "Route regex is not compiling: %s", k.Route)
			} else {
				log.Infof(server.Logger, "Stored wildcard rate limiting config for: %v", k)
				newWildcardMap[k] = &RateLimitingWildcardValue{
					RouteRegex:        routeRegex,
					RateLimitingValue: rateLimitingValue,
				}
			}
		}

		log.Infof(server.Logger, "Stored normal rate limiting config for: %v", k)
		newMap[k] = rateLimitingValue
	}

	// Log removed endpoints
	for k := range server.RateLimitingMap {
		if _, exists := newMap[k]; !exists {
			log.Infof(server.Logger, "Removed rate limiting entry as it is no longer part of the config: %v", k)
		}
	}

	// Replace both maps atomically
	server.RateLimitingMap = newMap
	server.RateLimitingWildcardMap = newWildcardMap
}

func Init(server *ServerData) {
	utils.StartPollingRoutine(server.PollingData.RateLimitingChannel, server.PollingData.RateLimitingTicker, AdvanceRateLimitingQueues, server)
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.RateLimitingChannel)
}

func isWildcardEndpoint(route string) bool {
	return strings.Contains(route, "*")
}
