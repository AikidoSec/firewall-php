package rate_limiting

import (
	. "main/aikido_types"
	"main/utils"
)

func AdvanceRateLimitingQueues(server *ServerData) {
	server.RateLimitingMutex.Lock()
	defer server.RateLimitingMutex.Unlock()

	for _, endpoint := range server.RateLimitingMap {
		AdvanceSlidingWindowMap(endpoint.UserCounts)
		AdvanceSlidingWindowMap(endpoint.IpCounts)
		AdvanceSlidingWindowMap(endpoint.RateLimitGroupCounts)
	}
}

func Init(server *ServerData) {
	utils.StartPollingRoutine(server.PollingData.RateLimitingChannel, server.PollingData.RateLimitingTicker, AdvanceRateLimitingQueues, server)
	AdvanceRateLimitingQueues(server)
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.RateLimitingChannel)
}
