package config

import (
	"encoding/json"
	"fmt"
	. "main/aikido_types"
	"main/globals"
	"main/log"
	"main/utils"
	"os"
)

func UpdateToken(token string) bool {
	if token == globals.CurrentToken {
		log.Debugf("Token is the same as previous one, skipping config reload...")
		return false
	}
	globals.CurrentToken = token
	globals.CurrentServer = globals.GetServer(token)
	log.Infof("Token changed to \"AIK_RUNTIME_***%s\"", utils.AnonymizeToken(token))
	return true
}

type ReloadResult int

const (
	ReloadError ReloadResult = iota
	ReloadWithSameToken
	ReloadWithNewToken
	ReloadWithPastSeenToken
)

func ReloadAikidoConfig(conf *AikidoConfigData, initJson string) ReloadResult {
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
		if !UpdateToken(conf.Token) {
			return ReloadWithSameToken
		}
		return ReloadWithPastSeenToken
	}
	server := globals.CreateServer(conf.Token)
	server.AikidoConfig = *conf
	UpdateToken(conf.Token)
	return ReloadWithNewToken
}

func Init(initJson string) {
	err := json.Unmarshal([]byte(initJson), &globals.EnvironmentConfig)
	if err != nil {
		panic(fmt.Sprintf("Error parsing JSON to EnvironmentConfig: %s", err))
	}

	globals.EnvironmentConfig.ServerPID = int32(os.Getppid())
	globals.EnvironmentConfig.RequestProcessorPID = int32(os.Getpid())

	conf := AikidoConfigData{}
	ReloadAikidoConfig(&conf, initJson)
	log.Init(conf.DiskLogs)
}

func Uninit() {
	log.Uninit()
}
