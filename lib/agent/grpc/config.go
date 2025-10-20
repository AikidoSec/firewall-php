package grpc

import (
	. "main/aikido_types"
	"main/ipc/protos"
	"main/log"
)

func storeConfig(server *ServerData, req *protos.Config) {
	server.AikidoConfig.ConfigMutex.Lock()
	defer server.AikidoConfig.ConfigMutex.Unlock()

	server.AikidoConfig.PlatformName = req.GetPlatformName()
	server.AikidoConfig.PlatformVersion = req.GetPlatformVersion()
	server.AikidoConfig.Endpoint = req.GetEndpoint()
	server.AikidoConfig.ConfigEndpoint = req.GetConfigEndpoint()
	server.AikidoConfig.Token = req.GetToken()
	server.AikidoConfig.LogLevel = req.GetLogLevel()
	server.AikidoConfig.DiskLogs = req.GetDiskLogs()
	server.AikidoConfig.Blocking = req.GetBlocking()
	server.AikidoConfig.LocalhostAllowedByDefault = req.GetLocalhostAllowedByDefault()
	server.AikidoConfig.CollectApiSchema = req.GetCollectApiSchema()

	log.SetLogLevel(server.AikidoConfig.LogLevel)
	log.Init(server.AikidoConfig.DiskLogs)
	log.Debugf("Updated Aikido Config with the one passed via gRPC!")
}
