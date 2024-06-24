package globals

const (
	Version                  = "1.0.0"
	ConfigFilePath           = "/opt/aikido/dev-config.json"
	LogFilePath              = "/var/log/aikido.log"
	SocketPath               = "/var/aikido.sock"
	ConfigAPIMethod          = "GET"
	ConfigAPI                = "/api/runtime/config"
	ConfigUpdatedAtMethod    = "GET"
	ConfigUpdatedAtAPI       = "/config"
	EventsAPIMethod          = "POST"
	EventsAPI                = "/api/runtime/events"
	MinHeartbeatIntervalInMS = 120000
)