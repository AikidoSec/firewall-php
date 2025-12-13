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

func UpdateToken(inst *instance.RequestProcessorInstance, token string) bool {
	server := globals.GetServer(token)
	if server == nil {
		log.Debugf(inst, "Server not found for token \"AIK_RUNTIME_***%s\"", utils.AnonymizeToken(token))
		return false
	}

	if token == inst.GetCurrentToken() {
		if inst.GetCurrentServer() == nil {
			inst.SetCurrentServer(server)
		}
		log.Debugf(inst, "Token is the same as previous one, skipping config reload...")
		return false
	}

	inst.SetCurrentToken(token)
	inst.SetCurrentServer(server)
	log.Infof(inst, "Token changed to \"AIK_RUNTIME_***%s\"", utils.AnonymizeToken(token))
	return true
}

type ReloadResult int

const (
	ReloadError ReloadResult = iota
	ReloadWithSameToken
	ReloadWithNewToken
	ReloadWithPastSeenToken
)

func ReloadAikidoConfig(inst *instance.RequestProcessorInstance, conf *AikidoConfigData, initJson string) ReloadResult {
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
		if !UpdateToken(inst, conf.Token) {
			return ReloadWithSameToken
		}
		return ReloadWithPastSeenToken
	}
	server := globals.CreateServer(conf.Token)
	server.AikidoConfig = *conf
	UpdateToken(inst, conf.Token)
	return ReloadWithNewToken
}

func Init(inst *instance.RequestProcessorInstance, initJson string) {
	err := json.Unmarshal([]byte(initJson), &globals.EnvironmentConfig)
	if err != nil {
		panic(fmt.Sprintf("Error parsing JSON to EnvironmentConfig: %s", err))
	}

	globals.EnvironmentConfig.ServerPID = int32(os.Getppid())
	globals.EnvironmentConfig.RequestProcessorPID = int32(os.Getpid())

	conf := AikidoConfigData{}
	ReloadAikidoConfig(inst, &conf, initJson)
	log.Init(conf.DiskLogs)
}

func Uninit() {
	log.Uninit()
}
