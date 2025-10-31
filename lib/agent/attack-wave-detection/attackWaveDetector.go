package attackwavedetection

import (
	. "main/aikido_types"
	"main/cloud"
	"main/ipc/protos"
	"main/log"
	"main/utils"
	"time"
)

// pushNewMinute pushes a new (zeroed) current-minute bucket,
// evicting the oldest if we exceed the window size, and fixing total accordingly.
func pushNewMinute(q *AttackWaveQueue) {
	q.Queue = append(q.Queue, 0)
	if len(q.Queue) > q.WindowSize {
		dropped := q.Queue[0]
		q.Queue = q.Queue[1:]
		q.Total -= dropped
	}
}

// incr increments the current minute bucket.
func incr(q *AttackWaveQueue) {
	if len(q.Queue) == 0 {
		pushNewMinute(q)
	}
	q.Queue[len(q.Queue)-1]++
	q.Total++
}

func sum(q *AttackWaveQueue) int { return q.Total }

func isEmpty(q *AttackWaveQueue) bool { return q.Total == 0 }

// deleteIpFromTracking removes an IP from all attack wave tracking structures.
func deleteIpFromTracking(server *ServerData, ip string) {
	delete(server.AttackWaveIpQueues, ip)
	delete(server.AttackWaveLastSeen, ip)
	delete(server.AttackWaveLastSent, ip)
}

func newAttackWaveQueue(windowSize int) *AttackWaveQueue {
	q := &AttackWaveQueue{
		WindowSize: windowSize,
		Queue:      make([]int, 0, windowSize),
	}
	pushNewMinute(q) // ensure there is a current bucket
	return q
}

func AdvanceAttackWaveQueues(server *ServerData) {
	server.AttackWaveMutex.Lock()
	defer server.AttackWaveMutex.Unlock()

	now := time.Now()

	if server.AttackWaveLastTick.IsZero() {
		server.AttackWaveLastTick = now
		return
	}

	elapsed := now.Sub(server.AttackWaveLastTick)
	tickDur := 1 * time.Minute
	steps := int(elapsed / tickDur)
	if steps <= 0 {
		return
	}
	if steps > server.AttackWaveWindowSize {
		steps = server.AttackWaveWindowSize
	}

	for i := 0; i < steps; i++ {
		for _, q := range server.AttackWaveIpQueues {
			pushNewMinute(q)
		}
		server.AttackWaveLastTick = server.AttackWaveLastTick.Add(tickDur)
	}

	// Light sweep: drop entries that are completely empty AND haven't been seen
	// in at least one full window interval.
	cutoff := now.Add(-time.Duration(server.AttackWaveWindowSize) * tickDur)
	for ip, q := range server.AttackWaveIpQueues {
		ls, ok := server.AttackWaveLastSeen[ip]
		if isEmpty(q) && (!ok || ls.Before(cutoff)) {
			deleteIpFromTracking(server, ip)
		}
	}
}

func Init(server *ServerData) {
	server.AttackWaveMutex.Lock()
	server.AttackWaveLastTick = time.Now()
	server.AttackWaveMutex.Unlock()

	utils.StartPollingRoutine(server.PollingData.AttackWaveChannel, server.PollingData.AttackWaveTicker, AdvanceAttackWaveQueues, server)
	AdvanceAttackWaveQueues(server)
}

func Uninit(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.AttackWaveChannel)
}

// IncrementAndDetect implements the detection logic using the sliding window:
// - skips if an event was recently sent for IP (minBetween)
// - increments current minute bucket for IP
// - if sum(window) >= threshold => mark event time and send event to cloud
func IncrementAndDetect(server *ServerData, ip string) bool {
	if ip == "" {
		return false
	}

	now := time.Now()

	server.AttackWaveMutex.Lock()
	defer server.AttackWaveMutex.Unlock()

	// throttle repeated events
	if last, ok := server.AttackWaveLastSent[ip]; ok && now.Sub(last) < server.AttackWaveMinBetween {
		// still update lastSeen so eviction stays fair
		server.AttackWaveLastSeen[ip] = now
		return false
	}

	// ensure queue exists
	q, ok := server.AttackWaveIpQueues[ip]
	if !ok {
		// capacity control (soft): if we exceed, evict the least-recently seen entry
		if len(server.AttackWaveIpQueues) >= server.AttackWaveMaxEntries {
			evictLRU(server)
		}
		q = newAttackWaveQueue(server.AttackWaveWindowSize)
		server.AttackWaveIpQueues[ip] = q
	}

	// increment for this request and mark lastSeen
	incr(q)
	server.AttackWaveLastSeen[ip] = now

	// check threshold within window
	if sum(q) < server.AttackWaveThreshold {
		return false
	}

	// threshold reached -> record event and return true
	server.AttackWaveLastSent[ip] = now
	if server.Logger != nil {
		log.Infof(server.Logger, "Attack wave detected from IP: %s", ip)
	}
	// report event to cloud
	cloud.SendAttackDetectedEvent(server, &protos.AttackDetected{
		Token:   server.AikidoConfig.Token,
		Request: &protos.Request{IpAddress: ip},
		Attack:  &protos.Attack{Metadata: []*protos.Metadata{}},
	}, "detected_attack_wave")

	return true
}

// evictLRU deletes the least-recently-seen IP (O(n), called only on cap pressure).
func evictLRU(server *ServerData) {
	var victim string
	var victimTS time.Time
	first := true
	for ip, ts := range server.AttackWaveLastSeen {
		if first || ts.Before(victimTS) {
			first = false
			victim = ip
			victimTS = ts
		}
	}
	// If lastSeen has no entries (shouldn't happen), fall back to arbitrary delete.
	if victim == "" {
		for ip := range server.AttackWaveIpQueues {
			victim = ip
			break
		}
	}
	if victim != "" {
		deleteIpFromTracking(server, victim)
	}
}
