package rate_limiting

import (
	. "main/aikido_types"
	"main/utils"
)

func AdvanceRateLimitingQueues(server *ServerData) {
	server.RateLimitingMutex.Lock()
	defer server.RateLimitingMutex.Unlock()

	for _, endpoint := range server.RateLimitingMap {
		AdvanceSlidingWindowMap(endpoint.UserCounts, endpoint.Config.WindowSizeInMinutes)
		AdvanceSlidingWindowMap(endpoint.IpCounts, endpoint.Config.WindowSizeInMinutes)
		AdvanceSlidingWindowMap(endpoint.RateLimitGroupCounts, endpoint.Config.WindowSizeInMinutes)
	}
}

func Init(server *ServerData) {
	utils.StartPollingRoutine(server.PollingData.RateLimitingChannel, server.PollingData.RateLimitingTicker, AdvanceRateLimitingQueues, server)
	AdvanceRateLimitingQueues(server)
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.RateLimitingChannel)
}
