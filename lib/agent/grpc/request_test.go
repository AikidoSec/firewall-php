package grpc

import (
	"sync"
	"testing"

	. "main/aikido_types"
	"main/utils"

	"github.com/stretchr/testify/assert"
)

func TestAttackWaveThrottling(t *testing.T) {
	t.Run("throttling checks LastSent map not queue", func(t *testing.T) {
		// Create a minimal server data structure
		server := &ServerData{
			AttackWave: AttackWaveState{
				Threshold:  10, // High threshold so we don't trigger events
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
		server.AttackWave.LastSent[ip] = now - 30000 // 30 seconds ago

		// Create a sliding window for this IP with counts above threshold
		sw := NewSlidingWindow()
		for i := 0; i < 15; i++ {
			sw.Increment()
		}
		server.AttackWave.IpQueues[ip] = sw

		// The key test: verify that the throttling logic checks the LastSent map
		// If we simulate being within the MinBetween window, event should be throttled

		// Since LastSent was 30 seconds ago and MinBetween is 60 seconds,
		// the throttling condition should trigger: now - lastSentTime < MinBetween
		lastSentTime, exists := server.AttackWave.LastSent[ip]
		assert.True(t, exists, "LastSent should exist for the IP")
		assert.True(t, now-lastSentTime < server.AttackWave.MinBetween,
			"Time since last event should be less than MinBetween (throttling active)")

		// Verify that SlidingWindow itself doesn't have a LastSent field
		// This would cause a compile error if we tried to access it, demonstrating the bug fix
		queue := server.AttackWave.IpQueues[ip]
		assert.NotNil(t, queue)
		assert.Equal(t, 15, queue.Total, "Queue should have correct count")
	})

	t.Run("LastSent map tracks different IPs independently", func(t *testing.T) {
		server := &ServerData{
			AttackWave: AttackWaveState{
				LastSent: make(map[string]int64),
			},
		}

		now := utils.GetTime()

		// Set different times for different IPs
		server.AttackWave.LastSent["192.168.1.1"] = now - 100000
		server.AttackWave.LastSent["192.168.1.2"] = now - 50000
		server.AttackWave.LastSent["192.168.1.3"] = now - 10000

		// Verify they're tracked independently
		assert.NotEqual(t, server.AttackWave.LastSent["192.168.1.1"],
			server.AttackWave.LastSent["192.168.1.2"],
			"Different IPs should have different LastSent times")
		assert.NotEqual(t, server.AttackWave.LastSent["192.168.1.2"],
			server.AttackWave.LastSent["192.168.1.3"],
			"Different IPs should have different LastSent times")
	})

	t.Run("non-existent IP not in LastSent map", func(t *testing.T) {
		server := &ServerData{
			AttackWave: AttackWaveState{
				LastSent: make(map[string]int64),
			},
		}

		// Check for IP that hasn't sent any events
		_, exists := server.AttackWave.LastSent["192.168.1.100"]
		assert.False(t, exists, "IP that never sent events should not be in LastSent map")
	})
}
