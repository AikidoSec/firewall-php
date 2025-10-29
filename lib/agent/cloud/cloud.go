package cloud

import (
	. "main/aikido_types"
	"main/utils"
)

func Init(server *ServerData) {
	server.StatsData.StartedAt = utils.GetTime()
	server.StatsData.MonitoredSinkTimings = make(map[string]MonitoredSinkTimings)

	utils.StartPollingRoutine(server.PollingData.HeartbeatRoutineChannel, server.PollingData.HeartbeatTicker, SendHeartbeatEvent, server)
	utils.StartPollingRoutine(server.PollingData.ConfigPollingRoutineChannel, server.PollingData.ConfigPollingTicker, CheckConfigUpdatedAt, server)
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.HeartbeatRoutineChannel)
	utils.StopPollingRoutine(server.PollingData.ConfigPollingRoutineChannel)
func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.HeartbeatRoutineChannel)
	utils.StopPollingRoutine(server.PollingData.ConfigPollingRoutineChannel)
}
