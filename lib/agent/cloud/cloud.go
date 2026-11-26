package cloud

import (
	. "main/aikido_types"
	"main/utils"
)

func Init(server *ServerData) {
	if !applyCachedConfig(server) {
		server.ConfigUpdateMutex.Lock()
		if !applyCachedConfig(server) {
			checkConfigUpdatedAt(server)
		}
		server.ConfigUpdateMutex.Unlock()
	}

	utils.StartPollingRoutine(server.PollingData.HeartbeatRoutineChannel, server.PollingData.HeartbeatTicker, SendHeartbeatEvent, server)
	utils.StartPollingRoutine(server.PollingData.ConfigPollingRoutineChannel, server.PollingData.ConfigPollingTicker, CheckConfigUpdatedAt, server)
}

func applyCachedConfig(server *ServerData) bool {
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()
	if server.CloudConfig.ConfigUpdatedAt == 0 {
		return false
	}
	ApplyCloudConfig(server)
	return true
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.HeartbeatRoutineChannel)
	utils.StopPollingRoutine(server.PollingData.ConfigPollingRoutineChannel)

	StopConfigStreamRoutine(server)
}
