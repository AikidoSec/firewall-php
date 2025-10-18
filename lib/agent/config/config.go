package config

import (
	"encoding/json"
	"fmt"
	"main/aikido_types"
	"main/globals"
	"main/log"
	"os"
)

func setConfigFromJson(jsonString []byte) bool {
	if err := json.Unmarshal(jsonString, &globals.EnvironmentConfig); err != nil {
		panic(fmt.Sprintf("Failed to unmarshal JSON to EnvironmentConfig: %v", err))
	}

	globals.InitialToken = os.Getenv("AIKIDO_TOKEN")
	if globals.InitialToken == "" {
		log.Infof("No token set! Aikido agent will load and wait for the token to be passed via gRPC!")
	}

	if globals.EnvironmentConfig.LogLevel != "" {
		if err := log.SetLogLevel(globals.EnvironmentConfig.LogLevel); err != nil {
			panic(fmt.Sprintf("Error setting log level: %s", err))
		}
	}

	if globals.EnvironmentConfig.SocketPath == "" {
		log.Errorf("No socket path set! Aikido agent will not load!")
		return false
	}

	return true
}

func Init(initJson string) bool {
	return setConfigFromJson([]byte(initJson))
}

func Uninit() {

}

func GetToken(server *aikido_types.ServerData) string {
	server.AikidoConfig.ConfigMutex.Lock()
	defer server.AikidoConfig.ConfigMutex.Unlock()

	return server.AikidoConfig.Token
}

func GetBlocking(server *aikido_types.ServerData) bool {
	server.AikidoConfig.ConfigMutex.Lock()
	defer server.AikidoConfig.ConfigMutex.Unlock()

	return server.AikidoConfig.Blocking
}
