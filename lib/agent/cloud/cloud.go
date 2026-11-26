package cloud

import (
	. "main/aikido_types"
	"main/globals"
	"main/utils"
	"sync"
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
	StopConfigStreamRoutine(server)
}

func PollCloudConfigs(_ *ServerData) {
	seen := make(map[*CloudPolicy]bool)
	var polls sync.WaitGroup
	for _, server := range globals.GetServers() {
		if seen[server.CloudPolicy] {
			continue
		}
		seen[server.CloudPolicy] = true
		polls.Go(func() { CheckConfigUpdatedAt(server) })
	}
	polls.Wait()
}
