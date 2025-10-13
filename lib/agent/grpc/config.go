package grpc

import (
	"main/aikido_types"
	"main/log"
)

func storeConfig(server *aikido_types.ServerData, token, logLevel string, diskLogs, blocking, localhostAllowedByDefault, collectApiSchema bool) {
	server.AikidoConfig.ConfigMutex.Lock()
	defer server.AikidoConfig.ConfigMutex.Unlock()

	if token != "" {
		server.AikidoConfig.Token = token
	}
	server.AikidoConfig.LogLevel = logLevel
	server.AikidoConfig.DiskLogs = diskLogs
	server.AikidoConfig.Blocking = blocking
	server.AikidoConfig.LocalhostAllowedByDefault = localhostAllowedByDefault
	server.AikidoConfig.CollectApiSchema = collectApiSchema

	log.SetLogLevel(server.AikidoConfig.LogLevel)
	log.Init(server)
	log.Debugf("Updated Aikido Config with the one passed via gRPC!")
}
