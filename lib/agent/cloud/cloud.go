package cloud

import (
	. "main/aikido_types"
	"main/globals"
	"main/utils"
	"sync"
)

func Init(server *ServerData) {
	server.ConfigRefreshMutex.Lock()
	server.CloudConfigMutex.Lock()
	cached := server.CloudConfig.ConfigUpdatedAt > 0
	if cached {
		ApplyCloudConfig(server)
	}
	server.CloudConfigMutex.Unlock()
	if !cached {
		CheckConfigUpdatedAt(server)
	}
	server.ConfigRefreshMutex.Unlock()

	utils.StartPollingRoutine(server.PollingData.HeartbeatRoutineChannel, server.PollingData.HeartbeatTicker, SendHeartbeatEvent, server)
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
		polls.Go(func() {
			server.ConfigRefreshMutex.Lock()
			defer server.ConfigRefreshMutex.Unlock()
			CheckConfigUpdatedAt(server)
		})
	}
	polls.Wait()
}
