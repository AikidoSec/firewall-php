package aikido_types

import (
	"regexp"

	"go4.org/netipx"
)

type EnvironmentConfigData struct {
	SocketPath                string `json:"socket_path"`                  // '/run/aikido-{version}/aikido-{datetime}-{randint}.sock'
	SAPI                      string `json:"sapi"`                         // '{php-sapi}'
	TrustProxy                bool   `json:"trust_proxy"`                  // default: true
	LocalhostAllowedByDefault bool   `json:"localhost_allowed_by_default"` // default: true
	CollectApiSchema          bool   `json:"collect_api_schema"`           // default: true
}

type AikidoConfigData struct {
	Token                     string `json:"token"`                        // default: ''
	LogLevel                  string `json:"log_level"`                    // default: 'WARN'
	Blocking                  bool   `json:"blocking"`                     // default: false
	TrustProxy                bool   `json:"trust_proxy"`                  // default: true
	LocalhostAllowedByDefault bool   `json:"localhost_allowed_by_default"` // default: true
	CollectApiSchema          bool   `json:"collect_api_schema"`           // default: true
	DiskLogs                  bool   `json:"disk_logs"`                    // default: false
}

type RateLimiting struct {
	Enabled        bool
	MaxRequests    int
	WindowSizeInMS int
}

type EndpointData struct {
	ForceProtectionOff bool
	RateLimiting       RateLimiting
	AllowedIPAddresses *netipx.IPSet
}

type WildcardEndpointData struct {
	RouteRegex *regexp.Regexp
	Data       EndpointData
}

type EndpointKey struct {
	Method string
	Route  string
}

type IpList struct {
	Description string
	IpSet       netipx.IPSet
}

type CloudConfigData struct {
	ConfigUpdatedAt   int64
	Endpoints         map[EndpointKey]EndpointData
	WildcardEndpoints map[string][]WildcardEndpointData
	BlockedUserIds    map[string]bool
	BypassedIps       *netipx.IPSet
	BlockedIps        map[string]IpList
	AllowedIps        map[string]IpList
	BlockedUserAgents *regexp.Regexp
	Block             int
}
