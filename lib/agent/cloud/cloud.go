package cloud

import (
	. "main/aikido_types"
	"main/globals"
	"main/utils"
	"time"
)

var (
	HeartbeatRoutineChannel     = make(chan struct{})
	HeartBeatTicker             = time.NewTicker(10 * time.Minute)
	ConfigPollingRoutineChannel = make(chan struct{})
	ConfigPollingTicker         = time.NewTicker(1 * time.Minute)
)

func Init() {
	for _, server := range globals.GetServers() {
		server.StatsData.StartedAt = utils.GetTime()
		server.StatsData.MonitoredSinkTimings = make(map[string]MonitoredSinkTimings)
		SendStartEvent(server)
	}
	utils.StartPollingRoutine(HeartbeatRoutineChannel, HeartBeatTicker, SendHeartbeatEvent)
	utils.StartPollingRoutine(ConfigPollingRoutineChannel, ConfigPollingTicker, CheckConfigUpdatedAt)
}

func Uninit() {
	utils.StopPollingRoutine(HeartbeatRoutineChannel)
	utils.StopPollingRoutine(ConfigPollingRoutineChannel)
}
