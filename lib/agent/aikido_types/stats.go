package aikido_types

import "sync"

type StatsDataType struct {
	StatsMutex sync.Mutex

	StartedAt       int64
	Requests        int
	RequestsAborted int
	Attacks         int
	AttacksBlocked  int

	MonitoredSinkStats map[string]MonitoredSinkStats
}

type RateLimitingConfig struct {
	MaxRequests         int
	WindowSizeInMinutes int
}

type RateLimitingCounts struct {
	NumberOfRequestsPerWindow Queue
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
