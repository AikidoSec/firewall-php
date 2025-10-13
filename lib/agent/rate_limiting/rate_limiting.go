package rate_limiting

import (
	"main/aikido_types"
	. "main/aikido_types"
	"main/globals"
	"main/log"
	"main/utils"
	"time"
)

var (
	RateLimitingChannel = make(chan struct{})
	RateLimitingTicker  = time.NewTicker(globals.MinRateLimitingIntervalInMs * time.Millisecond)
)

func advanceRateLimitingQueuesForMap(config *RateLimitingConfig, countsMap map[string]*RateLimitingCounts) {
	for _, counts := range countsMap {
		if config.WindowSizeInMinutes <= counts.NumberOfRequestsPerWindow.Length() {
			// Sliding window is moving, need to substract the entry that goes out of the window
			// Ex: if the window is set to 10 minutes, when another minute passes,
			//     need to remove the number of requests from the entry of 11 minutes ago

			// Get the number of requests for the entry that just dropped out of the sliding window
			numberOfRequestToSubstract := counts.NumberOfRequestsPerWindow.Pop()
			if counts.TotalNumberOfRequests < numberOfRequestToSubstract {
				// This should never happen, but better to have a check in place
				log.Warnf("More requests to substract (%d) than total number of requests (%d)!",
					numberOfRequestToSubstract, counts.TotalNumberOfRequests)
			} else {
				// Remove the number of requests for the entry that just dropped out of the sliding window from total
				counts.TotalNumberOfRequests -= numberOfRequestToSubstract
			}
		}

		// Create a new entry in queue for the current minute
		counts.NumberOfRequestsPerWindow.Push(0)
	}
}

func AdvanceRateLimitingQueues() {
	for _, server := range globals.Servers {
		server.RateLimitingMutex.Lock()
		defer server.RateLimitingMutex.Unlock()

		for _, endpoint := range server.RateLimitingMap {
			advanceRateLimitingQueuesForMap(&endpoint.Config, endpoint.UserCounts)
			advanceRateLimitingQueuesForMap(&endpoint.Config, endpoint.IpCounts)
			advanceRateLimitingQueuesForMap(&endpoint.Config, endpoint.RateLimitGroupCounts)
		}
	}
}

func Init(server *aikido_types.ServerData) {
	AdvanceRateLimitingQueues()
	utils.StartPollingRoutine(RateLimitingChannel, RateLimitingTicker, AdvanceRateLimitingQueues)
}

func Uninit() {
	utils.StopPollingRoutine(RateLimitingChannel)
}
