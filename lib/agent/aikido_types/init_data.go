package aikido_types

import (
	"sync"
	"time"
)

type MachineData struct {
	HostName   string `json:"hostname"`
	DomainName string `json:"domainname"`
	OS         string `json:"os"`
	OSVersion  string `json:"os_version"`
	IPAddress  string `json:"ip_address"`
}

type EnvironmentConfigData struct {
	SocketPath string `json:"socket_path"` // '/run/aikido-{version}/aikido-{datetime}-{randint}.sock'
	DiskLogs   bool   `json:"disk_logs"`   // default: false
	LogLevel   string `json:"log_level"`   // default: 'INFO'
}

type AikidoConfigData struct {
	ConfigMutex               sync.Mutex
	PlatformName              string `json:"platform_name"`                          // PHP platform name (fpm-fcgi, cli-server, ...)
	PlatformVersion           string `json:"platform_version"`                       // PHP version
	Token                     string `json:"token,omitempty"`                        // default: ''
	Endpoint                  string `json:"endpoint,omitempty"`                     // default: 'https://guard.aikido.dev/'
	ConfigEndpoint            string `json:"config_endpoint,omitempty"`              // default: 'https://runtime.aikido.dev/'
	LogLevel                  string `json:"log_level,omitempty"`                    // default: 'INFO'
	DiskLogs                  bool   `json:"disk_logs,omitempty"`                    // default: false
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
	AllowedIpsList        map[string]IpBlocklist
	BlockedUserAgents     string
	MonitoredIpsList      map[string]IpBlocklist
	MonitoredUserAgents   string
	UserAgentDetails      map[string]string
}

type IpsData struct {
	Key         string   `json:"key"`
	Source      string   `json:"source"`
	Description string   `json:"description"`
	Ips         []string `json:"ips"`
}

type UserAgentDetails struct {
	Key     string `json:"key"`
	Pattern string `json:"pattern"`
}

type ListsConfigData struct {
	Success              bool               `json:"success"`
	ServiceId            int                `json:"serviceId"`
	BlockedIpAddresses   []IpsData          `json:"blockedIPAddresses"`
	AllowedIpAddresses   []IpsData          `json:"allowedIPAddresses"`
	BlockedUserAgents    string             `json:"blockedUserAgents"`
	MonitoredIpAddresses []IpsData          `json:"monitoredIpAddresses"`
	MonitoredUserAgents  string             `json:"monitoredUserAgents"`
	UserAgentDetails     []UserAgentDetails `json:"userAgentDetails"`
}

type CloudConfigUpdatedAt struct {
	ServiceId       int   `json:"serviceId"`
	ConfigUpdatedAt int64 `json:"configUpdatedAt"`
}

type ServerDataPolling struct {
	HeartbeatRoutineChannel     chan struct{}
	HeartbeatTicker             *time.Ticker
	ConfigPollingRoutineChannel chan struct{}
	ConfigPollingTicker         *time.Ticker
	RateLimitingChannel         chan struct{}
	RateLimitingTicker          *time.Ticker
}

func NewServerDataPolling() *ServerDataPolling {
	return &ServerDataPolling{
		HeartbeatRoutineChannel:     make(chan struct{}),
		HeartbeatTicker:             time.NewTicker(10 * time.Minute),
		ConfigPollingRoutineChannel: make(chan struct{}),
		ConfigPollingTicker:         time.NewTicker(1 * time.Minute),
		RateLimitingChannel:         make(chan struct{}),
		RateLimitingTicker:          time.NewTicker(MinRateLimitingIntervalInMs * time.Millisecond),
	}
}

type ServerData struct {
	// Aikido config that contains info about endpoint, log_level, token, ...
	AikidoConfig AikidoConfigData

	// Cloud config that is obtain as a result from sending events to cloud or pulling the config when it changes
	CloudConfig CloudConfigData

	// Config mutex used to sync access to configuration data across the multiple go routines that we run in parallel
	CloudConfigMutex sync.Mutex

	// Polling data for the server, including mutex used to sync access to polling data across the go routines
	PollingData *ServerDataPolling

	// List of outgoing hostnames, their ports and number of hits, collected from the requests
	Hostnames      map[string]map[uint32]uint64
	HostnamesQueue Queue[string]

	// Hostnames mutex used to sync access to hostnames data across the go routines
	HostnamesMutex sync.Mutex

	// List of routes and their methods and count of calls collect from the requests
	// [method][route] = hits
	Routes      map[string]map[string]*Route
	RoutesQueue Queue[string]

	// Routes mutex used to sync access to routes data across the go routines
	RoutesMutex sync.Mutex

	// Global stats data, including mutex used to sync access to stats data across the go routines
	StatsData StatsDataType

	// Rate limiting map, which holds the current rate limiting state for each configured route
	// map[(method, route)] -> RateLimitingValue
	// method can also be '*'
	RateLimitingMap map[RateLimitingKey]*RateLimitingValue

	// Rate limiting wildcard map, which holds the current rate limiting state for each configured wildcard route
	// map[method] -> (RouteRegex, RateLimitingValue)
	// method can also be '*'
	RateLimitingWildcardMap map[RateLimitingKey]*RateLimitingWildcardValue

	// Rate limiting mutex used to sync access across the go routines
	RateLimitingMutex sync.RWMutex

	// Users map, which holds the current users and their data
	Users      map[string]User
	UsersQueue Queue[string]

	// Users mutex used to sync access across the go routines
	UsersMutex sync.Mutex

	// List of identified packages and their versions
	Packages map[string]Package

	// Packages mutex used to sync access to packages data across the go routines
	PackagesMutex sync.Mutex

	// MiddlewareInstalled boolean value to be reported on heartbeat events
	MiddlewareInstalled uint32

	// Got some request info passed via gRPC to the Agent
	GotTraffic uint32

	// Did we log a token error?
	LoggedTokenError uint32

	// Attacks detected events timestamps vector, used to limit the number of attacks reported to cloud
	AttackDetectedEventsSentAt []int64

	// Attack detected events timestamps vector mutex used to sync access across the go routines
	AttackDetectedEventsSentAtMutex sync.Mutex
}

const MaxNumberOfStoredHostnames = 2000
const MaxNumberOfStoredUsers = 2000
const MaxNumberOfStoredRoutes = 5000

func NewServerData() *ServerData {
	return &ServerData{
		Hostnames:               make(map[string]map[uint32]uint64),
		HostnamesQueue:          NewQueue[string](MaxNumberOfStoredHostnames),
		Routes:                  make(map[string]map[string]*Route),
		RoutesQueue:             NewQueue[string](MaxNumberOfStoredRoutes),
		RateLimitingMap:         make(map[RateLimitingKey]*RateLimitingValue),
		RateLimitingWildcardMap: make(map[RateLimitingKey]*RateLimitingWildcardValue),
		Users:                   make(map[string]User),
		UsersQueue:              NewQueue[string](MaxNumberOfStoredUsers),
		Packages:                make(map[string]Package),
		PollingData:             NewServerDataPolling(),
	}
}
