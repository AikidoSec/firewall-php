package globals

import (
	. "main/aikido_types"
	"sync"
)

var InitialToken = ""
var Servers = make(map[string]*ServerData)
var ServersMutex sync.RWMutex

func GetServer(token string) *ServerData {
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
