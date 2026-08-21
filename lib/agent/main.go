// Package main builds one executable used both as a short-lived launcher and as the
// long-lived shared Agent worker:
//
//	PHP -> Agent in launcher mode (no arguments) -> Agent in worker mode (--agent-worker)
//
// The launcher starts the worker and exits only after the shared Unix socket is ready.
// The worker retains the runtime-directory lock and continues independently of the PHP
// process that initiated startup. Both roles share one executable to avoid packaging
// a separate launcher.
package main

import (
	"errors"
	"fmt"
	. "main/aikido_types"
	"main/constants"
	"main/globals"
	"main/grpc"
	"main/log"
	"main/machine"
	"main/server_utils"
	"main/utils"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

var serversCleanupChannel = make(chan struct{})
var serversCleanupTicker = time.NewTicker(time.Minute)

const (
	agentWorkerArg       = "--agent-worker"
	agentStartupTimeout  = 10 * time.Second
	agentStartupInterval = 10 * time.Millisecond
	agentStartupAttempts = 2
)

func isProcessAlive(pid int32) bool {
	if pid <= 0 {
		return false
	}

	err := syscall.Kill(int(pid), 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

func serversCleanupRoutine(_ *ServerData) {
	for _, serverKey := range globals.GetServersKeys() {
		server := globals.GetServer(serverKey)
		if server == nil {
			continue
		}
		now := utils.GetTime()
		lastConnectionTime := atomic.LoadInt64(&server.LastConnectionTime)
		// Container platforms such as Google Cloud Run can suspend CPU between requests
		// while keeping PHP alive. Only unregister after the process exits.
		if now-lastConnectionTime > constants.MinServerInactivityForCleanup && !isProcessAlive(serverKey.ServerPID) {
			log.InfofMainAndServer(server.Logger, "Server \"AIK_RUNTIME_***%s\" (server PID: %d) has been inactive for more than 2 minutes, unregistering...", utils.AnonymizeToken(serverKey.Token), serverKey.ServerPID)
			server_utils.Unregister(serverKey)
		}
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
	log.Infof(log.MainLogger, "Aikido Agent v%s stopped!", constants.Version)
	log.Uninit()
}

func acquireAgentLock() (*os.File, error) {
	if err := os.MkdirAll(constants.RunPath, 0777); err != nil {
		return nil, err
	}

	runDirectory, err := os.Open(constants.RunPath)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(runDirectory.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		runDirectory.Close()
		return nil, err
	}

	return runDirectory, nil
}

func startAgentWorker() (<-chan error, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	command := exec.Command(executablePath, agentWorkerArg)
	command.Dir = "/"
	// Start the worker in its own session. Reparenting happens when this launcher
	// exits; setsid itself does not reparent the worker.
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := command.Start(); err != nil {
		return nil, err
	}

	workerExit := make(chan error, 1)
	// Reap candidates that exit during startup. If this candidate becomes the
	// long-lived worker, the launcher exits and the worker is reparented.
	go func() {
		workerExit <- command.Wait()
	}()
	return workerExit, nil
}

func isAgentReady() bool {
	connection, err := net.DialTimeout("unix", constants.SocketPath, agentStartupInterval)
	if err != nil {
		return false
	}
	connection.Close()
	return true
}

// runLauncher handles the public, no-argument invocation from the PHP extension.
// It starts a worker candidate when the Agent lock is free and waits for the
// shared socket. A failed candidate is reaped and retried.
func runLauncher() error {
	deadline := time.Now().Add(agentStartupTimeout)
	startupFailures := 0
	var workerExit <-chan error
	for {
		if isAgentReady() {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("Agent startup timed out after %s", agentStartupTimeout)
		}

		if workerExit != nil {
			select {
			case err := <-workerExit:
				workerExit = nil
				if err != nil {
					startupFailures++
					if startupFailures == agentStartupAttempts {
						return fmt.Errorf("Agent failed to become ready after %d startup attempts: %w", startupFailures, err)
					}
				}
			default:
				time.Sleep(agentStartupInterval)
				continue
			}
		}

		agentLock, err := acquireAgentLock()
		if err == nil {
			// Avoid spawning while a worker is active. Candidates take the lock
			// again themselves, so concurrent launchers can safely reach this point.
			agentLock.Close()
			workerExit, err = startAgentWorker()
			if err != nil {
				startupFailures++
				if startupFailures == agentStartupAttempts {
					return err
				}
			}
		} else if !errors.Is(err, syscall.EWOULDBLOCK) {
			return err
		}
		time.Sleep(agentStartupInterval)
	}
}

// runWorker handles only the launcher's private --agent-worker invocation. The
// first candidate to lock the runtime directory owns the gRPC socket and Agent
// services until it receives a termination signal. Other candidates exit
// successfully because another worker is already starting or running.
func runWorker() error {
	agentLock, err := acquireAgentLock()
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return nil
	}
	if err != nil {
		return err
	}
	defer agentLock.Close()

	if !AgentInit() {
		log.Errorf(log.MainLogger, "Agent initialization failed!")
		return fmt.Errorf("Agent initialization failed")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	signal := <-sigChan
	log.Infof(log.MainLogger, "Received signal: %s", signal)
	AgentUninit()
	return nil
}

// No arguments select the public launcher role. --agent-worker selects the
// private worker role and is accepted only with no additional arguments.
func runCommand(arguments []string) error {
	switch {
	case len(arguments) == 0:
		return runLauncher()
	case len(arguments) == 1 && arguments[0] == agentWorkerArg:
		return runWorker()
	default:
		return fmt.Errorf("unsupported Agent arguments: %v", arguments)
	}
}

func main() {
	if err := runCommand(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Aikido Agent failed: %v\n", err)
		os.Exit(-2)
	}
}
