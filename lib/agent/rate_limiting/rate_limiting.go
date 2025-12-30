package rate_limiting

import (
	. "main/aikido_types"
	"main/constants"
	"main/utils"
	"time"
)

func AdvanceRateLimitingQueues(server *ServerData) {
	server.RateLimitingMutex.RLock()
	endpoints := make([]*RateLimitingValue, 0, len(server.RateLimitingMap))
	for _, endpoint := range server.RateLimitingMap {
		endpoints = append(endpoints, endpoint)
	}
	server.RateLimitingMutex.RUnlock()

	for _, endpoint := range endpoints {
		endpoint.Mutex.Lock()
		AdvanceSlidingWindowMap(endpoint.UserCounts, endpoint.Config.WindowSizeInMinutes)
		AdvanceSlidingWindowMap(endpoint.IpCounts, endpoint.Config.WindowSizeInMinutes)
		AdvanceSlidingWindowMap(endpoint.RateLimitGroupCounts, endpoint.Config.WindowSizeInMinutes)
		endpoint.Mutex.Unlock()
	}
}

// StartRateLimitingTicker starts the rate limiting ticker
// Called on first request via sync.Once
func StartRateLimitingTicker(server *ServerData) {
	server.PollingData.RateLimitingTicker = time.NewTicker(constants.MinRateLimitingIntervalInMs * time.Millisecond)
	utils.StartPollingRoutine(server.PollingData.RateLimitingChannel, server.PollingData.RateLimitingTicker, AdvanceRateLimitingQueues, server)
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.RateLimitingChannel)
}
