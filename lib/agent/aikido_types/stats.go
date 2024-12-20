package aikido_types

import (
	"regexp"
	"sync"
)

type StatsDataType struct {
	StatsMutex sync.Mutex

	StartedAt       int64
	Requests        int
	RequestsAborted int
	Attacks         int
	AttacksBlocked  int
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

type RateLimitingWildcardValue struct {
	RouteRegex        *regexp.Regexp
	RateLimitingValue *RateLimitingValue
}
