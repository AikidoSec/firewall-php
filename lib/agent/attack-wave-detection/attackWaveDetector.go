package attack_wave_detection

import (
	. "main/aikido_types"
	"main/utils"
)

func AdvanceAttackWaveQueues(server *ServerData) {
	server.AttackWaveMutex.Lock()
	defer server.AttackWaveMutex.Unlock()

	AdvanceSlidingWindowMap(server.AttackWave.IpQueues, server.AttackWave.WindowSize)
	// remove entries from LastSent map if the time since the last event is greater than the MinBetween window
	now := utils.GetTime()
	for ip, lastSentTime := range server.AttackWave.LastSent {
		if now-lastSentTime > server.AttackWave.MinBetween {
			delete(server.AttackWave.LastSent, ip)
		}
	}
}

func Init(server *ServerData) {
	utils.StartPollingRoutine(server.PollingData.AttackWaveChannel, server.PollingData.AttackWaveTicker, AdvanceAttackWaveQueues, server)
	//[test] AdvanceAttackWaveQueues(server)
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.AttackWaveChannel)
}
