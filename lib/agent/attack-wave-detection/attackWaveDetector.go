package attackwavedetection

import (
	. "main/aikido_types"
	"main/utils"
)

func AdvanceAttackWaveQueues(server *ServerData) {
	server.AttackWaveMutex.Lock()
	defer server.AttackWaveMutex.Unlock()

	AdvanceSlidingWindowMap(server.AttackWaveIpQueues)
	//remove entries from AttackWaveLastSent and AttackWaveLastSeen if they are not in the AttackWaveIpQueues
	// TODO: find a better way to do this
	for ip := range server.AttackWaveLastSent {
		if _, ok := server.AttackWaveIpQueues[ip]; !ok {
			delete(server.AttackWaveLastSent, ip)
		}
	}
	for ip := range server.AttackWaveLastSeen {
		if _, ok := server.AttackWaveIpQueues[ip]; !ok {
			delete(server.AttackWaveLastSeen, ip)
		}
	}

}

func Init(server *ServerData) {
	utils.StartPollingRoutine(server.PollingData.AttackWaveChannel, server.PollingData.AttackWaveTicker, AdvanceAttackWaveQueues, server)
	AdvanceAttackWaveQueues(server)
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.AttackWaveChannel)
}
