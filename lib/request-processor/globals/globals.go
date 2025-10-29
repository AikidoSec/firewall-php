package globals

import (
	. "main/aikido_types"
	"sync"
)

var EnvironmentConfig EnvironmentConfigData
var Servers = make(map[string]*ServerData)
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
	}
}

func GetCurrentServer() *ServerData {
	return CurrentServer
}

func GetServer(token string) *ServerData {
	if token == "" {
		return nil
	}
	return Servers[token]
}

func GetServers() []*ServerData {
	servers := []*ServerData{}
	for _, server := range Servers {
		servers = append(servers, server)
	}
	return servers
}

func ServerExists(token string) bool {
	_, exists := Servers[token]
	return exists
}

func CreateServer(token string) *ServerData {
	Servers[token] = NewServerData()
	return Servers[token]
}

const (
	Version    = "1.4.2"
	SocketPath = "/run/aikido-" + Version + "/aikido-agent.sock"
)
