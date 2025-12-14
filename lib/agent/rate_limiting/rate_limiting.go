package rate_limiting

import (
	. "main/aikido_types"
	"main/utils"
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

func Init(server *ServerData) {
	utils.StartPollingRoutine(server.PollingData.RateLimitingChannel, server.PollingData.RateLimitingTicker, AdvanceRateLimitingQueues, server)
	AdvanceRateLimitingQueues(server)
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.RateLimitingChannel)
}
