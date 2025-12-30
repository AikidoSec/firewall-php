package cloud

import (
	. "main/aikido_types"
	"main/constants"
	"main/utils"
	"time"
)

func Init(server *ServerData) {
	server.StatsData.StartedAt = utils.GetTime()
	server.StatsData.MonitoredSinkTimings = make(map[string]MonitoredSinkTimings)

	CheckConfigUpdatedAt(server)

	// Start config polling immediately (for cloud config updates)
	// Heartbeat and other tickers will start on first request via StartAllTickers()
	utils.StartPollingRoutine(server.PollingData.ConfigPollingRoutineChannel, server.PollingData.ConfigPollingTicker, CheckConfigUpdatedAt, server)
}

// StartAllTickers starts all tickers on first request
// Called via sync.Once to ensure exactly-once execution, safe from any context
func StartAllTickers(server *ServerData) {
	// Determine initial heartbeat interval based on cloud config
	// Default to 10 minutes (conservative) if config was never fetched
	heartbeatInterval := 10 * time.Minute

	// Only use faster 1-minute interval if we successfully fetched config
	// and cloud indicates this is a new server (ReceivedAnyStats = false)
	if server.CloudConfig.ConfigUpdatedAt > 0 {
		if !server.CloudConfig.ReceivedAnyStats {
			heartbeatInterval = 1 * time.Minute
		} else if server.CloudConfig.HeartbeatIntervalInMS >= constants.MinHeartbeatIntervalInMS {
			heartbeatInterval = time.Duration(server.CloudConfig.HeartbeatIntervalInMS) * time.Millisecond
		}
	}

	server.PollingData.HeartbeatTicker = time.NewTicker(heartbeatInterval)
	utils.StartPollingRoutine(server.PollingData.HeartbeatRoutineChannel, server.PollingData.HeartbeatTicker, SendHeartbeatEvent, server)
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.HeartbeatRoutineChannel)
	utils.StopPollingRoutine(server.PollingData.ConfigPollingRoutineChannel)
}
