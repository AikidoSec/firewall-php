package grpc

import (
	"sync"
	"testing"

	. "main/aikido_types"
	"main/utils"

	"github.com/stretchr/testify/assert"
)

func TestAttackWaveThrottling(t *testing.T) {
	t.Run("returns false when event for IP was recently sent (within MinBetween window)", func(t *testing.T) {
		server := &ServerData{
			AttackWave: AttackWaveState{
				Threshold:  10,
				WindowSize: 20,
				MinBetween: 60000, // 60 seconds in milliseconds
				IpQueues:   make(map[string]*SlidingWindow),
				LastSent:   make(map[string]int64),
			},
			AttackWaveMutex: sync.Mutex{},
		}

		ip := "192.168.1.1"
		now := utils.GetTime()

		// Manually set LastSent time to recent past (30 seconds ago)
		server.AttackWave.LastSent[ip] = now - 30000

		// Create a sliding window for this IP with counts above threshold
		sw := NewSlidingWindow()
		for i := 0; i < 10; i++ {
			sw.Increment()
		}
		server.AttackWave.IpQueues[ip] = sw

		// Should return false (throttled) because last event was only 30 seconds ago (< 60s MinBetween)
		assert.False(t, updateAttackWaveCountsAndDetect(server, true, ip, "", ""))
	})

	t.Run("returns true and populates LastSent map when IP reaches threshold for first time", func(t *testing.T) {
		server := &ServerData{
			AttackWave: AttackWaveState{
				Threshold:  10,
				WindowSize: 20,
				MinBetween: 60000, // 60 seconds in milliseconds
				IpQueues:   make(map[string]*SlidingWindow),
				LastSent:   make(map[string]int64),
			},
			AttackWaveMutex: sync.Mutex{},
		}

		ip := "192.168.1.1"

		// Create a sliding window for this IP with counts above threshold
		sw := NewSlidingWindow()
		for i := 0; i < 9; i++ {
			sw.Increment()
		}
		server.AttackWave.IpQueues[ip] = sw

		// Verify IP doesn't exist in LastSent map initially
		_, exists := server.AttackWave.LastSent[ip]
		assert.False(t, exists, "IP should not be in LastSent map before threshold is reached")

		// Should return true (event sent) because this is the first time reaching threshold
		assert.True(t, updateAttackWaveCountsAndDetect(server, true, ip, "", ""))

		// Verify LastSent map was populated
		assert.True(t, server.AttackWave.LastSent[ip] > 0, "LastSent should be set after event is sent")
	})
}
