package globals

import (
	. "main/aikido_types"
	"sync"
)

var Machine MachineData

var Servers = make(map[ServerKey]*ServerData)
var PastDeletedServers = make(map[ServerKey]bool)
var ServersMutex sync.RWMutex

func GetServer(serverKey ServerKey) *ServerData {
	ServersMutex.RLock()
	defer ServersMutex.RUnlock()

	server, exists := Servers[serverKey]
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

func GetServersKeys() []ServerKey {
	ServersMutex.RLock()
	defer ServersMutex.RUnlock()

	serverKeys := []ServerKey{}
	for serverKey := range Servers {
		serverKeys = append(serverKeys, serverKey)
	}
	return serverKeys
}

func CreateServer(ServerKey ServerKey) *ServerData {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()
	Servers[ServerKey] = NewServerData()
	return Servers[ServerKey]
}

func GetOrCreateServer(key ServerKey, endpoint, configEndpoint string) (*ServerData, bool) {
	if server, exists := Servers[key]; exists {
		return server, true
	}
	server := NewServerData()
	for otherKey, other := range Servers {
		if otherKey.Token == key.Token && other.AikidoConfig.Endpoint == endpoint && other.AikidoConfig.ConfigEndpoint == configEndpoint {
			server.CloudPolicy = other.CloudPolicy
			break
		}
	}
	Servers[key] = server
	return server, false
}

func DeleteServer(key ServerKey) *ServerData {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()
	server := Servers[key]
	if server != nil {
		delete(Servers, key)
		PastDeletedServers[key] = true
	}
	return server
}

func IsPastDeletedServer(ServerKey ServerKey) bool {
	_, exists := PastDeletedServers[ServerKey]
	return exists
}
