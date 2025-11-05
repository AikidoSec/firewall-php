package attackwavedetection

import (
	. "main/aikido_types"
	"main/utils"
)

func AdvanceAttackWaveQueues(server *ServerData) {
	server.AttackWaveMutex.Lock()
	defer server.AttackWaveMutex.Unlock()

	AdvanceSlidingWindowMap(server.AttackWave.IpQueues, server.AttackWave.WindowSize)
}

func Init(server *ServerData) {
	utils.StartPollingRoutine(server.PollingData.AttackWaveChannel, server.PollingData.AttackWaveTicker, AdvanceAttackWaveQueues, server)
	AdvanceAttackWaveQueues(server)
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.AttackWaveChannel)
}
