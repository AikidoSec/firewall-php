package config

import (
	"encoding/json"
	"fmt"
	. "main/aikido_types"
	"main/globals"
	"main/log"
	"main/utils"
)

func UpdateToken(token string) {
	if token == globals.CurrentToken {
		log.Debugf("Token is the same as previous one, skipping config reload...")
		return
	}
	globals.CurrentToken = token
	globals.CurrentServer = globals.GetServer(token)
	log.Infof("Token changed to \"AIK_RUNTIME_***%s\"", utils.AnonymizeToken(token))
}

func ReloadAikidoConfig(conf *AikidoConfigData, initJson string) bool {
	err := json.Unmarshal([]byte(initJson), conf)
	if err != nil {
		return false
	}

	if err := log.SetLogLevel(conf.LogLevel); err != nil {
		return false
	}

	if conf.Token == "" {
		return false
	}

	if globals.ServerExists(conf.Token) {
		UpdateToken(conf.Token)
		return false
	}
	server := globals.CreateServer(conf.Token)
	server.AikidoConfig = *conf
	UpdateToken(conf.Token)
	return true
}

func Init(initJson string) {
	err := json.Unmarshal([]byte(initJson), &globals.EnvironmentConfig)
	if err != nil {
		panic(fmt.Sprintf("Error parsing JSON to EnvironmentConfig: %s", err))
	}

	conf := AikidoConfigData{}
	ReloadAikidoConfig(&conf, initJson)
	log.Init(conf.DiskLogs)
}

func Uninit() {
	log.Uninit()
}
