package config

import (
	"encoding/json"
	"fmt"
	"main/globals"
	"main/log"
	"os"
)

func setConfigFromJson(jsonString []byte) bool {
	if err := json.Unmarshal(jsonString, &globals.EnvironmentConfig); err != nil {
		panic(fmt.Sprintf("Failed to unmarshal JSON to EnvironmentConfig: %v", err))
	}

	if err := json.Unmarshal(jsonString, &globals.AikidoConfig); err != nil {
		panic(fmt.Sprintf("Failed to unmarshal JSON to AikidoConfig: %v", err))
	}

	if globals.AikidoConfig.LogLevel != "" {
		if err := log.SetLogLevel(globals.AikidoConfig.LogLevel); err != nil {
			panic(fmt.Sprintf("Error setting log level: %s", err))
		}
	}

	if globals.EnvironmentConfig.SocketPath == "" {
		log.Errorf("No socket path set! Aikido agent will not load!")
		return false
	}

	globals.AikidoConfig.Token = os.Getenv("AIKIDO_TOKEN")
	if globals.AikidoConfig.Token == "" {
		log.Infof("No token set! Aikido agent will load and wait for the token to be passed via gRPC!")
	}

	return true
}

func Init(initJson string) bool {
	return setConfigFromJson([]byte(initJson))
}

func Uninit() {

}

func GetToken() string {
	globals.AikidoConfig.ConfigMutex.Lock()
	defer globals.AikidoConfig.ConfigMutex.Unlock()

	return globals.AikidoConfig.Token
}

func GetBlocking() bool {
	globals.AikidoConfig.ConfigMutex.Lock()
	defer globals.AikidoConfig.ConfigMutex.Unlock()

	return globals.AikidoConfig.Blocking
}
