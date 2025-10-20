package attackwavedetection

import (
	"sync"
	"time"
)

// Sliding window queue (minute buckets)
// Each queue holds up to windowSize buckets, newest at the end.
type minuteQueue struct {
	windowSize int   // number of minutes in the window
	total      int   // running sum
	queue      []int // newest bucket at the end
}

func newMinuteQueue(windowSize int) *minuteQueue {
	mq := &minuteQueue{
		windowSize: windowSize,
		queue:      make([]int, 0, windowSize),
	}
	mq.pushNewMinute() // ensure there is a current bucket
	return mq
}

// pushNewMinute pushes a new (zeroed) current-minute bucket,
// evicting the oldest if we exceed the window size, and fixing total accordingly.
func (m *minuteQueue) pushNewMinute() {
	m.queue = append(m.queue, 0)
	if len(m.queue) > m.windowSize {
		dropped := m.queue[0]
		m.queue = m.queue[1:]
		m.total -= dropped
		if m.total < 0 {
			m.total = 0 // defensive; shouldn't happen
		}
	}
}

// incr increments the current minute bucket.
func (m *minuteQueue) incr() {
	if len(m.queue) == 0 {
		m.pushNewMinute()
	}
	m.queue[len(m.queue)-1]++
	m.total++
}

func (m *minuteQueue) sum() int { return m.total }

func (m *minuteQueue) isEmpty() bool { return m.total == 0 }

// ----------------------
// Options & Detector
// ----------------------
type Options struct {
	// detection
	AttackWaveThreshold  int           // default 15 (requests in window)
	WindowSizeInMinutes  int           // default 1 (matches TS timeframe of 60s)
	MinTimeBetweenEvents time.Duration // default 20m
	MaxEntries           int           // soft cap for IP maps; default 10000
	Tick                 time.Duration // default 1m; how often window advances
}

type AttackWaveDetector struct {
	// config
	attackWaveThreshold int
	windowSizeInMinutes int
	minBetween          time.Duration
	maxEntries          int

	// state
	mu            sync.Mutex
	ipQueues      map[string]*minuteQueue // per-IP sliding window
	lastSentEvent map[string]time.Time    // per-IP last event time (throttle)
	lastSeen      map[string]time.Time    // per-IP last request time (for eviction/sweep)

	// ticker
	tickDur  time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
	lastTick time.Time

	startOnce sync.Once
	stopOnce  sync.Once
}

// New creates the detector. Call Start() to begin advancing the window.
func New(opts Options) *AttackWaveDetector {
	// defaults
	if opts.AttackWaveThreshold <= 0 {
		opts.AttackWaveThreshold = 15
	}
	if opts.WindowSizeInMinutes <= 0 {
		opts.WindowSizeInMinutes = 1
	}
	if opts.MinTimeBetweenEvents <= 0 {
		opts.MinTimeBetweenEvents = 20 * time.Minute
	}
	if opts.MaxEntries <= 0 {
		opts.MaxEntries = 10000
	}
	if opts.Tick <= 0 {
		opts.Tick = time.Minute
	}

	return &AttackWaveDetector{
		attackWaveThreshold: opts.AttackWaveThreshold,
		windowSizeInMinutes: opts.WindowSizeInMinutes,
		minBetween:          opts.MinTimeBetweenEvents,
		maxEntries:          opts.MaxEntries,
		ipQueues:            make(map[string]*minuteQueue, 1024),
		lastSentEvent:       make(map[string]time.Time, 1024),
		lastSeen:            make(map[string]time.Time, 1024),
		tickDur:             opts.Tick,
		stopCh:              make(chan struct{}),
	}
}

// Start begins the background ticker that advances minute buckets.
// Safe to call multiple times; only starts once.
func (a *AttackWaveDetector) Start() {
	a.startOnce.Do(func() {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()

			now := time.Now()
			a.mu.Lock()
			a.lastTick = now
			a.mu.Unlock()

			t := time.NewTicker(a.tickDur)
			defer t.Stop()

			for {
				select {
				case tickTime := <-t.C:
					a.advanceByElapsed(tickTime)
				case <-a.stopCh:
					return
				}
			}
		}()
	})
}

// Stop terminates the advancing goroutine.
// Safe to call multiple times; only stops once.
func (a *AttackWaveDetector) Stop() {
	a.stopOnce.Do(func() {
		close(a.stopCh)
		a.wg.Wait()
	})
}

// Check implements the detection logic using the sliding window:
// - skips if an event was recently sent for IP (minBetween)
// - requires IsWebScanner(method, route, queryParams) to be true (pure pre-check)
// - increments current minute bucket for IP
// - if sum(window) >= threshold => mark event time and return true
func (a *AttackWaveDetector) Check(ip string, method string, route string, queryParams map[string]string) bool {
	if ip == "" {
		return false
	}
	// fast pre-check without lock
	if !IsWebScanner(method, route, queryParams) {
		return false
	}

	now := time.Now()

	a.mu.Lock()
	defer a.mu.Unlock()

	// throttle repeated events
	if last, ok := a.lastSentEvent[ip]; ok && now.Sub(last) < a.minBetween {
		// still update lastSeen so eviction stays fair
		a.lastSeen[ip] = now
		return false
	}

	// ensure queue exists
	q, ok := a.ipQueues[ip]
	if !ok {
		// capacity control (soft): if we exceed, evict the least-recently seen entry
		if len(a.ipQueues) >= a.maxEntries {
			a.evictLRU()
		}
		q = newMinuteQueue(a.windowSizeInMinutes)
		a.ipQueues[ip] = q
	}

	// increment for this request and mark lastSeen
	q.incr()
	a.lastSeen[ip] = now

	// check threshold within window
	if q.sum() < a.attackWaveThreshold {
		return false
	}

	// threshold reached -> record event and return true
	a.lastSentEvent[ip] = now
	return true
}

// advanceByElapsed advances the sliding window by one bucket per full tickDur elapsed,
// capped at windowSize to avoid huge catch-ups after long sleeps.
// Also performs a light sweep of idle IPs.
func (a *AttackWaveDetector) advanceByElapsed(now time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.lastTick.IsZero() {
		a.lastTick = now
		return
	}

	elapsed := now.Sub(a.lastTick)
	steps := int(elapsed / a.tickDur)
	if steps <= 0 {
		return
	}
	if steps > a.windowSizeInMinutes {
		steps = a.windowSizeInMinutes
	}

	for i := 0; i < steps; i++ {
		for _, q := range a.ipQueues {
			q.pushNewMinute()
		}
		a.lastTick = a.lastTick.Add(a.tickDur)
	}

	// Light sweep: drop entries that are completely empty AND haven't been seen
	// in at least one full window interval.
	cutoff := now.Add(-time.Duration(a.windowSizeInMinutes) * a.tickDur)
	for ip, q := range a.ipQueues {
		ls, ok := a.lastSeen[ip]
		if q.isEmpty() && (!ok || ls.Before(cutoff)) {
			delete(a.ipQueues, ip)
			delete(a.lastSeen, ip)
			delete(a.lastSentEvent, ip)
		}
	}
}

// evictLRU deletes the least-recently-seen IP (O(n), called only on cap pressure).
func (a *AttackWaveDetector) evictLRU() {
	var victim string
	var victimTS time.Time
	first := true
	for ip, ts := range a.lastSeen {
		if first || ts.Before(victimTS) {
			first = false
			victim = ip
			victimTS = ts
		}
	}
	// If lastSeen has no entries (shouldn't happen), fall back to arbitrary delete.
	if victim == "" {
		for ip := range a.ipQueues {
			victim = ip
			break
		}
	}
	if victim != "" {
		delete(a.ipQueues, victim)
		delete(a.lastSeen, victim)
		delete(a.lastSentEvent, victim)
	}
}
