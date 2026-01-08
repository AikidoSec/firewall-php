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

func InitializeServerLogger(server *ServerData, req *protos.Config) {
	storeConfig(server, req)
	serverKey := ServerKey{Token: req.GetToken(), ServerPID: req.GetServerPid()}
	server.Logger = log.CreateLogger(utils.AnonymizeToken(serverKey.Token), server.AikidoConfig.LogLevel, server.AikidoConfig.DiskLogs)
	atomic.StoreInt64(&server.LastConnectionTime, utils.GetTime())
}

func CompleteServerConfiguration(server *ServerData, serverKey ServerKey, req *protos.Config) {
	log.InfofMainAndServer(server.Logger, "Server \"AIK_RUNTIME_***%s\" (server PID: %d) registered successfully!", utils.AnonymizeToken(serverKey.Token), serverKey.ServerPID)

	cloud.Init(server)
	rate_limiting.StartRateLimitingTicker(server)
	attack_wave_detection.StartAttackWaveTicker(server)
	
	if globals.IsPastDeletedServer(serverKey) {
		log.InfofMainAndServer(server.Logger, "Server \"AIK_RUNTIME_***%s\" (server PID: %d) was registered before for this server PID, but deleted due to inactivity! Skipping start event as it was sent before...", utils.AnonymizeToken(serverKey.Token), serverKey.ServerPID)
	} else {
		cloud.SendStartEvent(server)
	}
}

func ConfigureServer(server *ServerData, req *protos.Config) {
	serverKey := ServerKey{Token: req.GetToken(), ServerPID: req.GetServerPid()}
	InitializeServerLogger(server, req)
	CompleteServerConfiguration(server, serverKey, req)
}

func Register(serverKey ServerKey, requestProcessorPID int32, req *protos.Config) {
	log.Infof(log.MainLogger, "Client (request processor PID: %d) connected. Registering server \"AIK_RUNTIME_***%s\" (server PID: %d)...", requestProcessorPID, utils.AnonymizeToken(serverKey.Token), serverKey.ServerPID)

	server := globals.CreateServer(serverKey)
	ConfigureServer(server, req)
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
