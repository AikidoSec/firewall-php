package main

import (
	"C"
	. "main/aikido_types"
	"main/globals"
	"main/grpc"
	"main/log"
	"main/machine"
	"main/server_utils"
	"main/utils"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)
import (
	"fmt"
	"main/constants"
)

var serversCleanupChannel = make(chan struct{})
var serversCleanupTicker = time.NewTicker(time.Minute)

func serversCleanupRoutine(_ *ServerData) {
	for _, serverKey := range globals.GetServersKeys() {
		server := globals.GetServer(serverKey)
		if server == nil {
			continue
		}
		now := utils.GetTime()
		lastConnectionTime := atomic.LoadInt64(&server.LastConnectionTime)
		if now-lastConnectionTime > constants.MinServerInactivityForCleanup {
			log.Infof(log.MainLogger, "Server \"AIK_RUNTIME_***%s\" (server PID: %d) has been inactive for more than 2 minutes, unregistering...", utils.AnonymizeToken(serverKey.Token), serverKey.ServerPID)
			server_utils.Unregister(serverKey)
		}
	}
}

func writePidFile() {
	pidFile, err := os.Create(constants.PidPath)
	if err != nil {
		log.Errorf(log.MainLogger, "Failed to create pid file: %v", err)
		return
	}
	defer pidFile.Close()
	pidFile.WriteString(fmt.Sprintf("%d", os.Getpid()))
}

func removePidFile() {
	if _, err := os.Stat(constants.PidPath); err == nil {
		os.Remove(constants.PidPath)
	}
}

func AgentInit() (initOk bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Warn(log.MainLogger, "Recovered from panic:", r)
			initOk = false
		}
	}()

	log.Init()
	machine.Init()
	if !grpc.Init() {
		return false
	}

	writePidFile()
	utils.StartPollingRoutine(serversCleanupChannel, serversCleanupTicker, serversCleanupRoutine, nil)

	log.Infof(log.MainLogger, "Aikido Agent v%s started!", constants.Version)
	return true
}

func AgentUninit() {
	utils.StopPollingRoutine(serversCleanupChannel)

	for _, serverKey := range globals.GetServersKeys() {
		server_utils.Unregister(serverKey)
	}
	grpc.Uninit()
	removePidFile()
	log.Infof(log.MainLogger, "Aikido Agent v%s stopped!", constants.Version)
	log.Uninit()
}

func main() {
	if !AgentInit() {
		log.Errorf(log.MainLogger, "Agent initialization failed!")
		os.Exit(-2)
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	signal := <-sigChan
	log.Infof(log.MainLogger, "Received signal: %s", signal)
	AgentUninit()
	os.Exit(0)
}
