package aikido_types

type OsInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Hostname struct {
	URL  string `json:"hostname"`
	Port int64  `json:"port,omitempty"`
}

type Route struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	Hits   int64  `json:"count"`
}

type User struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	LastIpAddress string `json:"lastIpAddress"`
	FirstSeenAt   int64  `json:"firstSeenAt"`
	LastSeenAt    int64  `json:"lastSeenAt"`
}

type AttacksDetected struct {
	Total   int `json:"total"`
	Blocked int `json:"blocked"`
}

type CompressedTiming struct {
	AverageInMS  float64            `json:"averageInMS"`
	Percentiles  map[string]float64 `json:"percentiles"`
	CompressedAt int64              `json:"compressedAt"`
}

type MonitoredSinkStats struct {
	AttacksDetected       AttacksDetected    `json:"attacksDetected"`
	InterceptorThrewError int                `json:"interceptorThrewError"`
	WithoutContext        int                `json:"withoutContext"`
	Total                 int                `json:"total"`
	CompressedTimings     []CompressedTiming `json:"compressedTimings"`
}

type Requests struct {
	Total           int             `json:"total"`
	Aborted         int             `json:"aborted"`
	AttacksDetected AttacksDetected `json:"attacksDetected"`
}

type Stats struct {
	Sinks     map[string]MonitoredSinkStats `json:"sinks"`
	StartedAt int64                         `json:"startedAt"`
	EndedAt   int64                         `json:"endedAt"`
	Requests  Requests                      `json:"requests"`
}

type AgentInfo struct {
	DryMode                   bool              `json:"dryMode"`
	Hostname                  string            `json:"hostname"`
	Version                   string            `json:"version"`
	IPAddress                 string            `json:"ipAddress"`
	OS                        OsInfo            `json:"os"`
	Packages                  map[string]string `json:"packages"`
	PreventPrototypePollution bool              `json:"preventedPrototypePollution"`
	NodeEnv                   string            `json:"nodeEnv"`
	Library                   string            `json:"library"`
}

type Started struct {
	Type  string    `json:"type"`
	Agent AgentInfo `json:"agent"`
	Time  int64     `json:"time"`
}

type Heartbeat struct {
	Type      string     `json:"type"`
	Stats     Stats      `json:"stats"`
	Hostnames []Hostname `json:"hostnames"`
	Routes    []Route    `json:"routes"`
	Users     []User     `json:"users"`
	Agent     AgentInfo  `json:"agent"`
	Time      int64      `json:"time"`
}
