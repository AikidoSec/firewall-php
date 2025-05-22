package aikido_types

import "sync"

type MonitoredSinkTimings struct {
	AttacksDetected       AttacksDetected
	InterceptorThrewError int
	WithoutContext        int
	Total                 int
	Timings               []int64
}

type StatsDataType struct {
	StatsMutex sync.Mutex

	StartedAt       int64
	Requests        int
	RequestsAborted int
	Attacks         int
	AttacksBlocked  int

	MonitoredSinkTimings map[string]MonitoredSinkTimings
}

type RateLimitingConfig struct {
	MaxRequests         int
	WindowSizeInMinutes int
}

type RateLimitingCounts struct {
	NumberOfRequestsPerWindow RateLimitingQueue
	TotalNumberOfRequests     int
}

type RateLimitingKey struct {
	Method string
	Route  string
}

type RateLimitingValue struct {
	Config     RateLimitingConfig
	UserCounts map[string]*RateLimitingCounts
	IpCounts   map[string]*RateLimitingCounts
}
