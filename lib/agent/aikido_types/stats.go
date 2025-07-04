package aikido_types

import (
	"regexp"
	"sync"
)

type MonitoredSinkTimings struct {
	Kind                  string
	AttacksDetected       AttacksDetected
	InterceptorThrewError int
	WithoutContext        int
	Total                 int
	Timings               []int64
	CompressedTimings     []CompressedTiming
}

type StatsDataType struct {
	StatsMutex sync.Mutex

	StartedAt           int64
	Requests            int
	RequestsAborted     int
	RequestsRateLimited int
	Attacks             int
	AttacksBlocked      int

	MonitoredSinkTimings map[string]MonitoredSinkTimings

	UserAgentsMatches  map[string]int
	IpAddressesMatches map[string]int
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
	Method     string
	Route      string
	Config     RateLimitingConfig
	UserCounts map[string]*RateLimitingCounts
	IpCounts   map[string]*RateLimitingCounts
}

type RateLimitingWildcardValue struct {
	RouteRegex        *regexp.Regexp
	RateLimitingValue *RateLimitingValue
}
