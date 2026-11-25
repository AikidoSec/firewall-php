package cloud

import (
	. "main/aikido_types"
	"main/utils"
	"slices"
)

func Init(server *ServerData) {
	policy := server.CloudPolicy
	policy.InitialSync.Lock()
	defer policy.InitialSync.Unlock()

	policy.CloudConfigMutex.Lock()
	policy.Servers = append(policy.Servers, server)
	if len(policy.Servers) == 1 {
		startConfigPolling(server)
	}
	cached := policy.CloudConfig.ConfigUpdatedAt > 0
	if cached {
		ApplyCloudConfig(server)
	}
	policy.CloudConfigMutex.Unlock()

	if !cached {
		CheckConfigUpdatedAt(server)
	}
	utils.StartPollingRoutine(server.PollingData.HeartbeatRoutineChannel, server.PollingData.HeartbeatTicker, SendHeartbeatEvent, server)
}

func startConfigPolling(server *ServerData) {
	utils.StartPollingRoutine(server.PollingData.ConfigPollingRoutineChannel, server.PollingData.ConfigPollingTicker, CheckConfigUpdatedAt, server)
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.HeartbeatRoutineChannel)
	policy := server.CloudPolicy
	policy.CloudConfigMutex.Lock()
	defer policy.CloudConfigMutex.Unlock()

	index := slices.Index(policy.Servers, server)
	policy.Servers = slices.Delete(policy.Servers, index, index+1)
	if index == 0 {
		utils.StopPollingRoutine(server.PollingData.ConfigPollingRoutineChannel)
		if len(policy.Servers) > 0 {
			startConfigPolling(policy.Servers[0])
		}
	}

	if policy.StreamOwner == server {
		StopConfigStreamRoutine(server)
		policy.StreamOwner = nil
		for _, other := range policy.Servers {
			startConfigStream(other)
			if policy.StreamOwner != nil {
				break
			}
		}
	}
}
