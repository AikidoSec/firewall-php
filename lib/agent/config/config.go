package config

import (
	. "main/aikido_types"
)

func GetToken(server *ServerData) string {
	server.AikidoConfig.ConfigMutex.Lock()
	defer server.AikidoConfig.ConfigMutex.Unlock()

	return server.AikidoConfig.Token
}

func GetBlocking(server *ServerData) bool {
	server.AikidoConfig.ConfigMutex.Lock()
	defer server.AikidoConfig.ConfigMutex.Unlock()

	return server.AikidoConfig.Blocking
}
