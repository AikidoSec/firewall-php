package aikido_types

import "sync"

type MachineData struct {
	HostName   string `json:"hostname"`
	DomainName string `json:"domainname"`
	OS         string `json:"os"`
	OSVersion  string `json:"os_version"`
	IPAddress  string `json:"ip_address"`
}

type EnvironmentConfigData struct {
	SocketPath      string `json:"socket_path"`               // '/run/aikido-{version}/aikido-{datetime}-{randint}.sock'
	PlatformName    string `json:"platform_name"`             // PHP platform name (fpm-fcgi, cli-server, ...)
	PlatformVersion string `json:"platform_version"`          // PHP version
	Endpoint        string `json:"endpoint,omitempty"`        // default: 'https://guard.aikido.dev/'
	ConfigEndpoint  string `json:"config_endpoint,omitempty"` // default: 'https://runtime.aikido.dev/'
	DiskLogs        bool   `json:"disk_logs,omitempty"`       // default: false
}

type AikidoConfigData struct {
	ConfigMutex               sync.Mutex
	Token                     string `json:"token,omitempty"`                        // default: ''
	LogLevel                  string `json:"log_level,omitempty"`                    // default: 'INFO'
	Blocking                  bool   `json:"blocking,omitempty"`                     // default: false
	LocalhostAllowedByDefault bool   `json:"localhost_allowed_by_default,omitempty"` // default: true
	CollectApiSchema          bool   `json:"collect_api_schema,omitempty"`           // default: true
}

type RateLimiting struct {
	Enabled        bool `json:"enabled"`
	MaxRequests    int  `json:"maxRequests"`
	WindowSizeInMS int  `json:"windowSizeInMS"`
}

type Endpoint struct {
	Method             string       `json:"method"`
	Route              string       `json:"route"`
	ForceProtectionOff bool         `json:"forceProtectionOff"`
	Graphql            interface{}  `json:"graphql"`
	AllowedIPAddresses []string     `json:"allowedIPAddresses"`
	RateLimiting       RateLimiting `json:"rateLimiting"`
}

type IpBlocklist struct {
	Description string
	Ips         []string
}

type CloudConfigData struct {
	Success               bool       `json:"success"`
	ServiceId             int        `json:"serviceId"`
	ConfigUpdatedAt       int64      `json:"configUpdatedAt"`
	HeartbeatIntervalInMS int        `json:"heartbeatIntervalInMS"`
	Endpoints             []Endpoint `json:"endpoints"`
	BlockedUserIds        []string   `json:"blockedUserIds"`
	BypassedIps           []string   `json:"allowedIPAddresses"`
	ReceivedAnyStats      bool       `json:"receivedAnyStats"`
	Block                 *bool      `json:"block,omitempty"`
	BlockedIpsList        map[string]IpBlocklist
	BlockedUserAgents     string
}

type BlockedIpsData struct {
	Source      string   `json:"source"`
	Description string   `json:"description"`
	Ips         []string `json:"ips"`
}

type ListsConfigData struct {
	Success            bool             `json:"success"`
	ServiceId          int              `json:"serviceId"`
	BlockedIpAddresses []BlockedIpsData `json:"blockedIPAddresses"`
	BlockedUserAgents  string           `json:"blockedUserAgents"`
}

type CloudConfigUpdatedAt struct {
	ServiceId       int   `json:"serviceId"`
	ConfigUpdatedAt int64 `json:"configUpdatedAt"`
}
