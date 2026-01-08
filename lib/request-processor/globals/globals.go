package globals

import (
	. "main/aikido_types"
	"regexp"
	"sync"
)

var EnvironmentConfig EnvironmentConfigData
var Servers = make(map[string]*ServerData)
var ServersMutex sync.RWMutex
var CurrentToken string = ""
var CurrentServer *ServerData = nil

func NewServerData() *ServerData {
	return &ServerData{
		AikidoConfig: AikidoConfigData{},
		CloudConfig: CloudConfigData{
			Block: -1,
		},
		CloudConfigMutex:    sync.Mutex{},
		MiddlewareInstalled: false,
		ParamMatchers:       make(map[string]*regexp.Regexp),
	}
}

func GetCurrentServer() *ServerData {
	return CurrentServer
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
	Version    = "1.4.13"
	SocketPath = "/run/aikido-" + Version + "/aikido-agent.sock"
)
