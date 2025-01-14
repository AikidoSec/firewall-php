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
	SendStartEvent()
	utils.StartPollingRoutine(HeartbeatRoutineChannel, HeartBeatTicker, SendHeartbeatEvent)
	utils.StartPollingRoutine(ConfigPollingRoutineChannel, ConfigPollingTicker, CheckConfigUpdatedAt)

	globals.StatsData.StartedAt = utils.GetTime()
	globals.StatsData.MonitoredSinkTimings = make(map[string]MonitoredSinkTimings)
	globals.MiddlewareInstalled = 0
}

func Uninit() {
	utils.StopPollingRoutine(HeartbeatRoutineChannel)
	utils.StopPollingRoutine(ConfigPollingRoutineChannel)
}
