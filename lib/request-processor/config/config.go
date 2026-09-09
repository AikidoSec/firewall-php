package config

import (
	"encoding/json"
	. "main/aikido_types"
	"main/globals"
	"main/grpc"
	"main/instance"
	"main/log"
	"main/utils"
	"os"
	"time"
)

func UpdateToken(instance *instance.RequestProcessorInstance, token string) bool {
	if instance.GetCurrentToken() == token {
		log.Debugf(instance, "Token is the same as previous one, skipping config reload...")
		return false
	}

	server := globals.GetServer(token)
	if server == nil {
		log.Debugf(instance, "Server not found for token \"AIK_RUNTIME_***%s\"", utils.AnonymizeToken(token))
		return false
	}
	instance.SetCurrentToken(token)
	instance.SetCurrentServer(server)
	log.Infof(instance, "Token changed to \"AIK_RUNTIME_***%s\"", utils.AnonymizeToken(token))
	return true
}

func ReloadAikidoConfig(instance *instance.RequestProcessorInstance, conf *AikidoConfigData, initJson string) bool {
	err := json.Unmarshal([]byte(initJson), conf)
	if err != nil {
		return false
	}

	// Invalid logging config must not prevent switching to the site's server.
	_ = log.SetLogLevel(conf.LogLevel)

	if conf.Token == "" {
		// A tokenless site must not retain the previous site's policy or reporting token.
		instance.SetCurrentToken("")
		instance.SetCurrentServer(nil)
		return true
	}

	if !globals.ServerExists(conf.Token) {
		server := globals.CreateServer(conf.Token)
		server.AikidoConfig = *conf
	}
	if !UpdateToken(instance, conf.Token) {
		return true
	}

	initializeServer(instance.GetCurrentServer())
	return true
}

func initializeServer(server *ServerData) {
	server.ServerInitMutex.Lock()
	if !server.ServerInitialized {
		grpc.SendAikidoConfig(server)
		grpc.OnPackages(server, server.AikidoConfig.Packages)
		server.ServerInitialized = true
	}
	server.ServerInitMutex.Unlock()
	grpc.GetCloudConfig(server, 5*time.Second)
}

func Init(platformName string, serverPID int32) {
	globals.EnvironmentConfig.ServerPID = serverPID
	globals.EnvironmentConfig.RequestProcessorPID = int32(os.Getpid())
	globals.EnvironmentConfig.PlatformName = platformName
}

func InitInstance(instance *instance.RequestProcessorInstance, initJson string) {
	conf := AikidoConfigData{}
	ReloadAikidoConfig(instance, &conf, initJson)
	log.Init(conf.DiskLogs)
}

func Uninit() {
	log.Uninit()
}
