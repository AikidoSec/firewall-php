package config

import (
	"encoding/json"
	. "main/aikido_types"
	"main/globals"
	"main/instance"
	"main/log"
	"main/utils"
	"os"
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

type ReloadResult int

const (
	ReloadError ReloadResult = iota
	ReloadWithSameToken
	ReloadWithNewToken
	ReloadWithPastSeenToken
)

func ReloadAikidoConfig(instance *instance.RequestProcessorInstance, conf *AikidoConfigData, initJson string) ReloadResult {
	err := json.Unmarshal([]byte(initJson), conf)
	if err != nil {
		return ReloadError
	}

	if err := log.SetLogLevel(conf.LogLevel); err != nil {
		return ReloadError
	}

	if conf.Token == "" {
		return ReloadError
	}

	if globals.ServerExists(conf.Token) {
		if !UpdateToken(instance, conf.Token) {
			return ReloadWithSameToken
		}
		return ReloadWithPastSeenToken
	}
	server := globals.CreateServer(conf.Token)
	server.AikidoConfig = *conf
	UpdateToken(instance, conf.Token)
	return ReloadWithNewToken
}

func getServerPID(platformName string) int32 {
	switch platformName {
	case "fpm-fcgi", "apache2handler":
		// Requests run in child processes that share a stable server master.
		return int32(os.Getppid())
	case "cli-server", "frankenphp":
		// Requests run in the server process itself.
		return int32(os.Getpid())
	default:
		return int32(os.Getpid())
	}
}

func Init(platformName string) {
	globals.EnvironmentConfig.ServerPID = getServerPID(platformName)
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
