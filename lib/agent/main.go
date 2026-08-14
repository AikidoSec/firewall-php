// Package main builds one executable that is used in two roles:
//
//	PHP --posix_spawn--> launcher (no arguments)
//	                           `--start--> worker (--agent-worker)
//
// The Agent must keep running after PHP module startup and serve every PHP
// process using the same runtime directory. PHP cannot wait for that long-lived
// process because module startup would never finish. If PHP started it without
// waiting, each supported PHP host would instead be responsible for reaping it
// when it eventually exits.
//
// The short-lived launcher avoids both outcomes. PHP waits for and reaps only
// the launcher. The launcher locks the versioned runtime directory, starts the
// long-lived worker with that lock already inherited, and exits. The worker is
// then reparented to init or a subreaper, which becomes responsible for reaping
// it. The worker retains the lock, owns the PID file and Unix socket, and serves
// all PHP processes using that runtime directory.
//
// These roles share one executable only to avoid packaging and versioning a
// separate launcher binary. They are separate processes with separate lifetimes.
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
	"os/exec"
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

const (
	agentWorkerArg = "--agent-worker"
	// The launcher opens and locks the runtime directory using whatever descriptor
	// its process receives. ExtraFiles maps that same open directory to descriptor
	// 3 in the worker. The worker keeps its copy after the launcher exits, so the
	// lock is never released; reopening the directory would leave an unlocked gap.
	agentLockFD = 3
)

func serversCleanupRoutine(_ *ServerData) {
	for _, serverKey := range globals.GetServersKeys() {
		server := globals.GetServer(serverKey)
		if server == nil {
			continue
		}
		now := utils.GetTime()
		lastConnectionTime := atomic.LoadInt64(&server.LastConnectionTime)
		if now-lastConnectionTime > constants.MinServerInactivityForCleanup {
			log.InfofMainAndServer(server.Logger, "Server \"AIK_RUNTIME_***%s\" (server PID: %d) has been inactive for more than 2 minutes, unregistering...", utils.AnonymizeToken(serverKey.Token), serverKey.ServerPID)
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

func acquireAgentLock() (*os.File, error) {
	if err := os.MkdirAll(constants.RunPath, 0777); err != nil {
		return nil, err
	}
	if err := os.Chmod(constants.RunPath, 0777); err != nil {
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

// getInheritedAgentLock validates the directory descriptor passed by the launcher.
func getInheritedAgentLock() (*os.File, error) {
	runDirectory := os.NewFile(agentLockFD, constants.RunPath)
	if runDirectory == nil {
		return nil, fmt.Errorf("missing inherited Agent lock")
	}

	lockInfo, err := runDirectory.Stat()
	if err != nil {
		runDirectory.Close()
		return nil, err
	}
	runDirectoryInfo, err := os.Stat(constants.RunPath)
	if err != nil || !os.SameFile(lockInfo, runDirectoryInfo) {
		runDirectory.Close()
		return nil, fmt.Errorf("invalid inherited Agent lock")
	}
	if err := syscall.Flock(agentLockFD, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		runDirectory.Close()
		return nil, err
	}

	return runDirectory, nil
}

// runLauncher handles the public, no-argument invocation from the PHP extension.
// It either finds that another process owns Agent startup or starts one worker
// while keeping the runtime-directory lock held across the process handoff.
func runLauncher() error {
	agentLock, err := acquireAgentLock()
	if err == syscall.EWOULDBLOCK {
		return nil
	}
	if err != nil {
		return err
	}
	defer agentLock.Close()

	executablePath, err := os.Executable()
	if err != nil {
		return err
	}
	nullDevice, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer nullDevice.Close()

	command := exec.Command(executablePath, agentWorkerArg)
	command.Dir = "/"
	command.Stdin = nullDevice
	command.Stdout = nullDevice
	command.Stderr = nullDevice
	// Start the worker in its own session. Reparenting happens when this launcher
	// exits; setsid itself does not reparent the worker.
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	// Pass the locked directory as the worker's agentLockFD.
	command.ExtraFiles = []*os.File{agentLock}
	if err := command.Start(); err != nil {
		return err
	}
	// Release only Go's process handle. The worker keeps running after this
	// launcher exits and is then adopted by init or a subreaper.
	return command.Process.Release()
}

// runWorker handles only the launcher's private --agent-worker invocation. It
// retains the inherited lock for its lifetime and owns the PID file, gRPC socket,
// and Agent services until it receives a termination signal.
func runWorker() error {
	agentLock, err := getInheritedAgentLock()
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
