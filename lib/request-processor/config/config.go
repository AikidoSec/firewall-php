package config

import (
	"encoding/json"
	"fmt"
	. "main/aikido_types"
	"main/globals"
	"main/instance"
	"main/log"
	"main/utils"
	"os"
)

func UpdateToken(instance *instance.RequestProcessorInstance, token string) bool {
	server := globals.GetServer(token)
	if token == instance.GetCurrentToken() {
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

func Init(instance *instance.RequestProcessorInstance, initJson string) {
	err := json.Unmarshal([]byte(initJson), &globals.EnvironmentConfig)
	if err != nil {
		panic(fmt.Sprintf("Error parsing JSON to EnvironmentConfig: %s", err))
	}

	globals.EnvironmentConfig.ServerPID = int32(os.Getppid())
	globals.EnvironmentConfig.RequestProcessorPID = int32(os.Getpid())

	conf := AikidoConfigData{}
	ReloadAikidoConfig(instance, &conf, initJson)
	log.Init(conf.DiskLogs)
}

func Uninit() {
	log.Uninit()
}
