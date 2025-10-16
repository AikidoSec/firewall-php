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
	tmpEnvironmentConfigData := aikido_types.EnvironmentConfigData{}
	if err := json.Unmarshal(jsonString, &tmpEnvironmentConfigData); err != nil {
		panic(fmt.Sprintf("Failed to unmarshal JSON to EnvironmentConfig: %v", err))
	}

	globals.InitialToken = os.Getenv("AIKIDO_TOKEN")
	if globals.InitialToken == "" {
		log.Infof("No token set! Aikido agent will load and wait for the token to be passed via gRPC!")
	}

	globals.ServersMutex.Lock()
	globals.Servers[globals.InitialToken] = aikido_types.NewServerData()
	initialServer := globals.Servers[globals.InitialToken]
	globals.ServersMutex.Unlock()

	initialServer.EnvironmentConfig = tmpEnvironmentConfigData

	if err := json.Unmarshal(jsonString, &initialServer.AikidoConfig); err != nil {
		panic(fmt.Sprintf("Failed to unmarshal JSON to AikidoConfig: %v", err))
	}
	initialServer.AikidoConfig.Token = globals.InitialToken

	if initialServer.AikidoConfig.LogLevel != "" {
		if err := log.SetLogLevel(initialServer.AikidoConfig.LogLevel); err != nil {
			panic(fmt.Sprintf("Error setting log level: %s", err))
		}
	}

	if initialServer.EnvironmentConfig.SocketPath == "" {
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
