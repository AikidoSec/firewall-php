package rate_limiting

import (
	. "main/aikido_types"
	"main/utils"
	"time"
)

// NextResetAt returns the next tick on this server's minute timer.
func NextResetAt(server *ServerData, now time.Time) time.Time {
	return server.RateLimitingStartedAt.Add((now.Sub(server.RateLimitingStartedAt)/time.Minute + 1) * time.Minute)
}

func advanceRateLimitingQueues(server *ServerData, nextReset time.Time) {
	server.RateLimitingMutex.RLock()
	// Config updates replace the whole map, so we can iterate this map after unlocking.
	endpoints := server.RateLimitingMap
	server.RateLimitingMutex.RUnlock()

	for _, endpoint := range endpoints {
		endpoint.Mutex.Lock()
		// An endpoint created after this tick already belongs to the new minute.
		if endpoint.NextResetAt.Before(nextReset) {
			AdvanceSlidingWindowMap(endpoint.UserCounts, endpoint.Config.WindowSizeInMinutes)
			AdvanceSlidingWindowMap(endpoint.IpCounts, endpoint.Config.WindowSizeInMinutes)
			AdvanceSlidingWindowMap(endpoint.RateLimitGroupCounts, endpoint.Config.WindowSizeInMinutes)
			endpoint.NextResetAt = nextReset
		}
		endpoint.Mutex.Unlock()
	}
}

func AdvanceRateLimitingQueues(server *ServerData) {
	advanceRateLimitingQueues(server, NextResetAt(server, time.Now()))
}

func Init(server *ServerData) {
	server.RateLimitingStartedAt = time.Now()
	server.PollingData.RateLimitingTicker.Reset(time.Minute)
	utils.StartPollingRoutine(server.PollingData.RateLimitingChannel, server.PollingData.RateLimitingTicker, AdvanceRateLimitingQueues, server)
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.RateLimitingChannel)
}
