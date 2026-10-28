package attack_wave_detection

import (
	. "main/aikido_types"
	"main/utils"
	"sync"
)

func advanceAttackWaveState(state *AttackWaveState, mutex *sync.Mutex) {
	mutex.Lock()
	defer mutex.Unlock()

	AdvanceSlidingWindowMap(state.IpQueues, state.WindowSize)
	// remove entries from LastSent map if the time since the last event is greater than the MinBetween window
	now := utils.GetTime()
	for ip, lastSentTime := range state.LastSent {
		if now-lastSentTime > state.MinBetween {
			delete(state.LastSent, ip)
		}
	}
}

func AdvanceAttackWaveQueues(server *ServerData) {
	advanceAttackWaveState(&server.AttackWave, &server.AttackWaveMutex)
	advanceAttackWaveState(&server.RuntimeAnomalyWave, &server.RuntimeAnomalyWaveMutex)
}

func Init(server *ServerData) {
	utils.StartPollingRoutine(server.PollingData.AttackWaveChannel, server.PollingData.AttackWaveTicker, AdvanceAttackWaveQueues, server)
	AdvanceAttackWaveQueues(server)
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.AttackWaveChannel)
}
