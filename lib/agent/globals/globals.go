package globals

import (
	. "main/aikido_types"
)

var Machine MachineData

var Servers = make(map[string]*ServerData)

func GetServer(token string) *ServerData {
	server, exists := Servers[token]
	if !exists {
		return nil
	}
	return server
}

func GetServers() []*ServerData {
	servers := []*ServerData{}
	for _, server := range Servers {
		servers = append(servers, server)
	}
	return servers
}

func GetServersTokens() []string {
	tokens := []string{}
	for token := range Servers {
		tokens = append(tokens, token)
	}
	return tokens
}

func CreateServer(token string) *ServerData {
	Servers[token] = NewServerData()
	return Servers[token]
}

func DeleteServer(token string) {
	delete(Servers, token)
}
