package aikido_types

import (
	"regexp"

	"go4.org/netipx"
)

type EnvironmentConfigData struct {
	PlatformName        string `json:"platform_name"` // PHP platform name (fpm-fcgi, cli-server, ...)
	ServerPID           int32
	RequestProcessorPID int32
}

type AikidoConfigData struct {
	PlatformName              string            `json:"platform_name"`                // PHP platform name (fpm-fcgi, cli-server, ...)
	PlatformVersion           string            `json:"platform_version"`             // PHP version
	Endpoint                  string            `json:"endpoint"`                     // default: 'https://guard.aikido.dev/'
	ConfigEndpoint            string            `json:"config_endpoint"`              // default: 'https://runtime.aikido.dev/'
	Token                     string            `json:"token"`                        // default: ''
	LogLevel                  string            `json:"log_level"`                    // default: 'WARN'
	Blocking                  bool              `json:"blocking"`                     // default: false
	TrustProxy                bool              `json:"trust_proxy"`                  // default: true
	LocalhostAllowedByDefault bool              `json:"localhost_allowed_by_default"` // default: true
	CollectApiSchema          bool              `json:"collect_api_schema"`           // default: true
	DiskLogs                  bool              `json:"disk_logs"`                    // default: false
	Packages                  map[string]string `json:"packages"`                     // default: {}
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
	Key         string
	Description string
	IpSet       netipx.IPSet
}

type CloudConfigData struct {
	ConfigUpdatedAt          int64
	Endpoints                map[EndpointKey]EndpointData
	WildcardEndpoints        map[string][]WildcardEndpointData
	BlockedUserIds           map[string]bool
	BypassedIps              *netipx.IPSet
	BlockedIps               map[string]IpList
	AllowedIps               map[string]IpList
	BlockedUserAgents        *regexp.Regexp
	MonitoredIps             map[string]IpList
	MonitoredUserAgents      *regexp.Regexp
	UserAgentDetails         map[string]*regexp.Regexp
	Block                    int
	BlockNewOutgoingRequests bool
	OutboundDomains          map[string]string // hostname -> mode (allow/block)
}
