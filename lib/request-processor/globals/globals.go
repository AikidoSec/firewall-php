package globals

import (
	"log"
	"os"
	"regexp"
	"sync"

	. "main/aikido_types"
)

var EnvironmentConfig EnvironmentConfigData
var Servers = make(map[string]*ServerData)
var ServersMutex sync.RWMutex

type LogLevel int

const (
	LogDebugLevel LogLevel = iota
	LogInfoLevel
	LogWarnLevel
	LogErrorLevel
)

var (
	CurrentLogLevel = LogErrorLevel
	Logger          = log.New(os.Stdout, "", 0)
	CliLogging      = true
	LogFilePath     = ""
	LogMutex        sync.RWMutex
	LogFile         *os.File
)

func NewServerData() *ServerData {
	return &ServerData{
		AikidoConfig: AikidoConfigData{},
		CloudConfig: CloudConfigData{
			Block: -1,
		},
		CloudConfigMutex:    sync.Mutex{},
		MiddlewareInstalled: false,
		ServerInitialized:   false,
		ServerInitMutex:     sync.Mutex{},
		ParamMatchers:       make(map[string]*regexp.Regexp),
	}
}

func GetServer(token string) *ServerData {
	if token == "" {
		return nil
	}
	ServersMutex.RLock()
	defer ServersMutex.RUnlock()
	return Servers[token]
}

func GetServers() []*ServerData {
	ServersMutex.RLock()
	defer ServersMutex.RUnlock()
	servers := []*ServerData{}
	for _, server := range Servers {
		servers = append(servers, server)
	}
	return servers
}

func ServerExists(token string) bool {
	ServersMutex.RLock()
	defer ServersMutex.RUnlock()
	_, exists := Servers[token]
	return exists
}

func CreateServer(token string) *ServerData {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()
	Servers[token] = NewServerData()
	return Servers[token]
}

const (
	Version    = "1.5.6"
	SocketPath = "/run/aikido-" + Version + "/aikido-agent.sock"
)

var SocketPath = "/run/aikido-" + Version + "/aikido-agent.sock"

// SetRuntimeDir switches the socket directory to /tmp when running on Lambda.
// The flag is passed from C++ (via RequestProcessorInit) because Go's
// os.LookupEnv is unreliable when this shared library is loaded into a
// forked FPM worker.
func SetRuntimeDir(isLambda bool) {
	if isLambda {
		SocketPath = "/tmp/aikido-" + Version + "/aikido-agent.sock"
	}
}
