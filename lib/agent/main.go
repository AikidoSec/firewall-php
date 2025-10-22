package main

import (
	"C"
	"main/config"
	"main/globals"
	"main/grpc"
	"main/log"
	"main/machine"
)
import (
	. "main/aikido_types"
	"main/server_utils"
	"main/utils"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

var serversCleanupChannel = make(chan struct{})
var serversCleanupTicker = time.NewTicker(2 * time.Minute)

func serversCleanupRoutine(_ *ServerData) {
	for _, token := range globals.GetServersTokens() {
		server := globals.GetServer(token)
		if server == nil {
			continue
		}
		now := utils.GetTime()
		lastConnectionTime := atomic.LoadInt64(&server.LastConnectionTime)
		if now-lastConnectionTime > MinServerInactivityForCleanup {
			// Server has been inactive
			log.Infof("Server has been inactive for more than 2 minutes, unregistering...")
			server_utils.Unregister(token)
		}
	}
}

func AgentInit(initJson string) (initOk bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Warn("Recovered from panic:", r)
			initOk = false
		}
	}()

	if !config.Init(initJson) {
		return false
	}

	log.Init(globals.EnvironmentConfig.DiskLogs)
	log.Infof("Loaded local config: %+v", globals.EnvironmentConfig)

	machine.Init()
	if !grpc.Init(globals.EnvironmentConfig.SocketPath) {
		return false
	}

	utils.StartPollingRoutine(serversCleanupChannel, serversCleanupTicker, serversCleanupRoutine, nil)

	log.Infof("Aikido Agent v%s started!", Version)
	return true
}

func AgentUninit() {
	utils.StopPollingRoutine(serversCleanupChannel)

	for _, token := range globals.GetServersTokens() {
		server_utils.Unregister(token)
	}
	grpc.Uninit()
	config.Uninit()

	log.Infof("Aikido Agent v%s stopped!", Version)
	log.Uninit()
}

func main() {
	if len(os.Args) != 2 {
		log.Errorf("Usage: %s <init_json>", os.Args[0])
		os.Exit(-1)
	}
	if !AgentInit(os.Args[1]) {
		log.Errorf("Agent initialization failed!")
		os.Exit(-2)
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	signal := <-sigChan
	log.Infof("Received signal: %s", signal)
	AgentUninit()
	os.Exit(0)
}
