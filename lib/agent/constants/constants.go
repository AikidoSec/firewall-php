package constants

const (
	Version                             = "1.5.22"
	RunPath                             = "/run/aikido-" + Version
	SocketPath                          = RunPath + "/aikido-agent.sock"
	ConfigUpdatedAtMethod               = "GET"
	ConfigUpdatedAtAPI                  = "/config"
	ConfigAPIMethod                     = "GET"
	ConfigAPI                           = "/api/runtime/config"
	ConfigStreamMethod                  = "GET"
	ConfigStreamAPI                     = "/api/runtime/stream"
	ConfigUpdatedEvent                  = "config-updated"
	ConfigStreamInitialReconnectInMs    = 5000      // delay before the first attempt to reconnect a dropped config stream
	ConfigStreamMaxReconnectInMs        = 60 * 1000 // upper bound of the exponential reconnect backoff
	ConfigStreamStableConnectionInMs    = 30 * 1000 // a connection that lasted at least this long resets the backoff
	ConfigStreamReadTimeoutInMs         = 70 * 1000 // no data received for this long means the connection is considered dead
	ConfigStreamDialTimeoutInMs         = 30 * 1000 // upper bound for establishing the connection of the config stream
	ConfigStreamTLSHandshakeTimeoutInMs = 10 * 1000 // upper bound for the TLS handshake of the config stream
	CloudRequestTimeoutInMs             = 30 * 1000 // upper bound for a cloud request, so a stuck request cannot block a routine forever
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

	// Name under which the cloud enables the config stream for a service
	RealtimeUpdatesFeature = "realtime_updates"
)
