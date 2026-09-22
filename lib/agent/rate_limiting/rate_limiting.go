package rate_limiting

import (
	. "main/aikido_types"
	"main/utils"
	"time"
)

func advanceRateLimitingQueues(server *ServerData, tick, now time.Time) {
	server.RateLimitingMutex.Lock()
	defer server.RateLimitingMutex.Unlock()
	// Tickers can drop ticks when processing is delayed. Keep the next reset
	// aligned with the next scheduled tick, rather than the callback's finish time.
	server.RateLimitingNextResetAt = tick.Add((now.Sub(tick)/time.Minute + 1) * time.Minute)

	for _, endpoint := range server.RateLimitingMap {
		endpoint.Mutex.Lock()
		defer endpoint.Mutex.Unlock()
		AdvanceSlidingWindowMap(endpoint.UserCounts, endpoint.Config.WindowSizeInMinutes)
		AdvanceSlidingWindowMap(endpoint.IpCounts, endpoint.Config.WindowSizeInMinutes)
		AdvanceSlidingWindowMap(endpoint.RateLimitGroupCounts, endpoint.Config.WindowSizeInMinutes)
	}
}

func Init(server *ServerData) {
	ticker := server.PollingData.RateLimitingTicker
	now := time.Now()
	ticker.Reset(time.Minute)
	advanceRateLimitingQueues(server, now, now)
	go func() {
		for {
			select {
			case tick := <-ticker.C:
				advanceRateLimitingQueues(server, tick, time.Now())
			case <-server.PollingData.RateLimitingChannel:
				ticker.Stop()
				return
			}
		}
	}()
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.RateLimitingChannel)
}
