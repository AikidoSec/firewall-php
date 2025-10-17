package cloud

import (
	. "main/aikido_types"
	"main/utils"
)

func InitServer(server *ServerData) {
	server.StatsData.StartedAt = utils.GetTime()
	server.StatsData.MonitoredSinkTimings = make(map[string]MonitoredSinkTimings)
	SendStartEvent(server)

	utils.StartPollingRoutine(server.PollingData.HeartbeatRoutineChannel, server.PollingData.HeartBeatTicker, SendHeartbeatEvent, server)
	utils.StartPollingRoutine(server.PollingData.ConfigPollingRoutineChannel, server.PollingData.ConfigPollingTicker, CheckConfigUpdatedAt, server)
}

func UninitServer(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.HeartbeatRoutineChannel)
	utils.StopPollingRoutine(server.PollingData.ConfigPollingRoutineChannel)
}
