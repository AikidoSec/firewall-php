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
	"main/cloud"
	"main/rate_limiting"
	"os"
	"os/signal"
	"syscall"
)

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
	log.Init()
	log.Infof("Loaded local config: %+v", globals.EnvironmentConfig)

	machine.Init()
	if !grpc.Init() {
		return false
	}

	cloud.Init()
	rate_limiting.Init()

	log.Infof("Aikido Agent v%s started!", globals.Version)
	return true
}

func AgentUninit() {
	rate_limiting.Uninit()
	cloud.Uninit()
	grpc.Uninit()
	config.Uninit()

	log.Infof("Aikido Agent v%s stopped!", globals.Version)
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
