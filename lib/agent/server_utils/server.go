package server_utils

import (
	. "main/aikido_types"
	"main/cloud"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"main/rate_limiting"
	"main/utils"
	"sync/atomic"
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
}

func Register(token string, req *protos.Config) {
	log.Infof(log.MainLogger, "Registering server %s...", utils.AnonymizeToken(token))

	server := globals.CreateServer(token)
	storeConfig(server, req)
	server.Logger = log.CreateLogger(utils.AnonymizeToken(token), server.AikidoConfig.LogLevel, server.AikidoConfig.DiskLogs)

	atomic.StoreInt64(&server.LastConnectionTime, utils.GetTime())

	cloud.Init(server)
	rate_limiting.Init(server)

	log.Infof(log.MainLogger, "Server %s registered successfully!", utils.AnonymizeToken(token))
}

func Unregister(token string) {
	log.Infof(log.MainLogger, "Unregistering server %s...", utils.AnonymizeToken(token))
	server := globals.GetServer(token)
	if server == nil {
		return
	}
	rate_limiting.Uninit(server)
	cloud.Uninit(server)
	globals.DeleteServer(token)

	log.Infof(log.MainLogger, "Server %s unregistered successfully!", utils.AnonymizeToken(token))
}
