package globals

import (
	. "main/aikido_types"
	"sync"
)

var EnvironmentConfig EnvironmentConfigData
var Servers = make(map[string]*ServerData)
var ServersMutex sync.RWMutex
var CurrentToken string

func NewServerData() *ServerData {
	return &ServerData{
		AikidoConfig: AikidoConfigData{},
		CloudConfig: CloudConfigData{
			Block: -1,
		},
		CloudConfigMutex:    sync.Mutex{},
		MiddlewareInstalled: false,
	}
}

func GetCurrentServer() *ServerData {
	return GetServer(CurrentToken)
}

func GetServer(token string) *ServerData {
	ServersMutex.RLock()
	defer ServersMutex.RUnlock()
	if token == "" {
		return nil
	}
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
	Version = "1.4.0"
)
