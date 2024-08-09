package globals

const (
	Version                     = "1.0.41"
	SocketPath                  = "/run/aikido-" + Version + ".sock"
	ConfigAPIMethod             = "GET"
	ConfigAPI                   = "/api/runtime/config"
	ConfigUpdatedAtMethod       = "GET"
	ConfigUpdatedAtAPI          = "/config"
	EventsAPIMethod             = "POST"
	EventsAPI                   = "/api/runtime/events"
	MinHeartbeatIntervalInMS    = 120000
	MinRateLimitingIntervalInMs = 60000   // 1 minute
	MaxRateLimitingIntervalInMs = 3600000 // 1 hour
)
