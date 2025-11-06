package server_utils

import (
	. "main/aikido_types"
	attack_wave_detection "main/attack-wave-detection"
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

func Register(serverKey ServerKey, requestProcessorPID int32, req *protos.Config) {
	log.Infof(log.MainLogger, "Client (request processor PID: %d) connected. Registering server \"AIK_RUNTIME_***%s\" (server PID: %d)...", requestProcessorPID, utils.AnonymizeToken(serverKey.Token), serverKey.ServerPID)

	server := globals.CreateServer(serverKey)
	storeConfig(server, req)
	server.Logger = log.CreateLogger(utils.AnonymizeToken(serverKey.Token), server.AikidoConfig.LogLevel, server.AikidoConfig.DiskLogs)

	log.Infof(server.Logger, "Server \"AIK_RUNTIME_***%s\" (server PID: %d) registered successfully!", utils.AnonymizeToken(serverKey.Token), serverKey.ServerPID)

	atomic.StoreInt64(&server.LastConnectionTime, utils.GetTime())

	cloud.Init(server)
	if globals.IsPastDeletedServer(serverKey) {
		log.Infof(server.Logger, "Server \"AIK_RUNTIME_***%s\" (server PID: %d) was registered before for this server PID, but deleted due to inactivity! Skipping start event as it was sent before...", utils.AnonymizeToken(serverKey.Token), serverKey.ServerPID)
	} else {
		cloud.SendStartEvent(server)
	}

	rate_limiting.Init(server)
	attack_wave_detection.Init(server)

	log.Infof(log.MainLogger, "Server \"AIK_RUNTIME_***%s\" (server PID: %d) registered successfully!", utils.AnonymizeToken(serverKey.Token), serverKey.ServerPID)
}

func Unregister(serverKey ServerKey) {
	log.Infof(log.MainLogger, "Unregistering server \"AIK_RUNTIME_***%s\" (server PID: %d)...", utils.AnonymizeToken(serverKey.Token), serverKey.ServerPID)
	server := globals.GetServer(serverKey)
	if server == nil {
		return
	}
	attack_wave_detection.Uninit(server)
	rate_limiting.Uninit(server)
	cloud.Uninit(server)
	globals.DeleteServer(serverKey)

	log.Infof(log.MainLogger, "Server \"AIK_RUNTIME_***%s\" (server PID: %d) unregistered successfully!", utils.AnonymizeToken(serverKey.Token), serverKey.ServerPID)
}
