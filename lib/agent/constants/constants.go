package constants

const (
	Version                             = "1.4.7"
	SocketPath                          = "/run/aikido-" + Version + "/aikido-agent.sock"
	PidPath                             = "/run/aikido-" + Version + "/aikido-agent.pid"
	ConfigUpdatedAtMethod               = "GET"
	ConfigUpdatedAtAPI                  = "/config"
	ConfigAPIMethod                     = "GET"
	ConfigAPI                           = "/api/runtime/config"
	ListsAPIMethod                      = "GET"
	ListsAPI                            = "api/runtime/firewall/lists"
	EventsAPIMethod                     = "POST"
	EventsAPI                           = "/api/runtime/events"
	MinHeartbeatIntervalInMS            = 120000
	MinRateLimitingIntervalInMs         = 60000   // 1 minute
	MaxRateLimitingIntervalInMs         = 3600000 // 1 hour
	MaxAttackDetectedEventsPerInterval  = 100
	AttackDetectedEventsIntervalInMs    = 60 * 60 * 1000 // 1 hour
	MinStatsCollectedForRelevantMetrics = 1000
	MinServerInactivityForCleanup       = 2 * 60 * 1000 // 2 minutes - time interval for checking if registered servers are inactive (they are not running anymore), so the Agent can cleanup their memory
	MaxSlidingWindowEntries             = 100000        // max number of entries in the sliding window
)
