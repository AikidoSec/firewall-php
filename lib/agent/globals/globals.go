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

func DeleteServer(ServerKey ServerKey) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()
	delete(Servers, ServerKey)
	PastDeletedServers[ServerKey] = true
}

func IsPastDeletedServer(ServerKey ServerKey) bool {
	ServersMutex.RLock()
	defer ServersMutex.RUnlock()
	_, exists := PastDeletedServers[ServerKey]
	return exists
}
