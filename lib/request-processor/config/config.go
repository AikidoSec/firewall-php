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

func UpdateToken(token string) {
	if token == globals.CurrentToken {
		log.Debugf("Token is the same as previous one, skipping config reload...")
		return
	}
	globals.CurrentToken = token
	globals.CurrentServer = globals.GetServer(token)
	log.Infof("Token changed to \"AIK_RUNTIME_***%s\"", utils.AnonymizeToken(token))
}

func ReloadAikidoConfig(conf *AikidoConfigData, initJson string) {
	err := json.Unmarshal([]byte(initJson), conf)
	if err != nil {
		panic(fmt.Sprintf("Error parsing JSON to AikidoConfig: %s", err))
	}

	if err := log.SetLogLevel(conf.LogLevel); err != nil {
		panic(fmt.Sprintf("Error setting log level: %s", err))
	}

	if conf.Token != "" {
		server := globals.CreateServer(conf.Token)
		server.AikidoConfig = *conf
		UpdateToken(conf.Token)
	}
}

func Init(initJson string) {
	err := json.Unmarshal([]byte(initJson), &globals.EnvironmentConfig)
	if err != nil {
		panic(fmt.Sprintf("Error parsing JSON to EnvironmentConfig: %s", err))
	}

	globals.EnvironmentConfig.RequestProcessorPID = int32(os.Getpid())
	globals.EnvironmentConfig.ServerPID = int32(os.Getppid())

	conf := AikidoConfigData{}
	ReloadAikidoConfig(&conf, initJson)
	log.Init(conf.DiskLogs)
}

func Uninit() {
	log.Uninit()
}
