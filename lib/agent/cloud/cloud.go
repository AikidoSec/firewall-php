package cloud

import (
	. "main/aikido_types"
	"main/utils"
)

func Init(server *ServerData) {
	if !applyCachedConfig(server) {
		server.ConfigRefreshMutex.Lock()
		if !applyCachedConfig(server) {
			checkConfigUpdatedAt(server)
		}
		server.ConfigRefreshMutex.Unlock()
	}

	utils.StartPollingRoutine(server.PollingData.HeartbeatRoutineChannel, server.PollingData.HeartbeatTicker, SendHeartbeatEvent, server)
	utils.StartPollingRoutine(server.PollingData.ConfigPollingRoutineChannel, server.PollingData.ConfigPollingTicker, CheckConfigUpdatedAt, server)
}

func applyCachedConfig(server *ServerData) bool {
	server.ConfigCacheMutex.Lock()
	defer server.ConfigCacheMutex.Unlock()
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
