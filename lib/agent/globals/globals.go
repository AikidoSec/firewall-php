package globals

import (
	. "main/aikido_types"
	"sync"
)

var Machine MachineData

var EnvironmentConfig EnvironmentConfigData

var Servers = make(map[string]*ServerData)
var ServersMutex sync.RWMutex

func GetServer(token string) *ServerData {
	ServersMutex.RLock()
	defer ServersMutex.RUnlock()

	server, exists := Servers[token]
	if !exists {
		return nil
	}
	return server
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

func GetServersTokens() []string {
	ServersMutex.RLock()
	defer ServersMutex.RUnlock()

	tokens := []string{}
	for token := range Servers {
		tokens = append(tokens, token)
	}
	return tokens
}

func CreateServer(token string) *ServerData {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()
	Servers[token] = NewServerData()
	return Servers[token]
}

func DeleteServer(token string) *ServerData {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()
	delete(Servers, token)
	return Servers[token]
}
