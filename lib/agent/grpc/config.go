package grpc

import (
	"main/globals"
	"main/log"
)

func storeConfig(token, logLevel string, diskLogs, blocking, localhostAllowedByDefault, collectApiSchema bool) {
	globals.AikidoConfig.ConfigMutex.Lock()
	defer globals.AikidoConfig.ConfigMutex.Unlock()

	if token != "" {
		globals.AikidoConfig.Token = token
	}
	globals.AikidoConfig.LogLevel = logLevel
	globals.AikidoConfig.DiskLogs = diskLogs
	globals.AikidoConfig.Blocking = blocking
	globals.AikidoConfig.LocalhostAllowedByDefault = localhostAllowedByDefault
	globals.AikidoConfig.CollectApiSchema = collectApiSchema

	log.SetLogLevel(globals.AikidoConfig.LogLevel)
	log.Init()
	log.Debugf("Updated Aikido Config with the one passed via gRPC!")
}
