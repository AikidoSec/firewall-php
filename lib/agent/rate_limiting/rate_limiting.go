package rate_limiting

import (
	. "main/aikido_types"
	"main/utils"
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

func Init(server *ServerData) {
	utils.StartPollingRoutine(server.PollingData.RateLimitingChannel, server.PollingData.RateLimitingTicker, AdvanceRateLimitingQueues, server)
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.RateLimitingChannel)
}
