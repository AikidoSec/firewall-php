// Package main builds the Aikido Agent executable.
//
// One installed executable is invoked in two process roles:
//
//	PHP --posix_spawn--> launcher (no arguments)
//	                           `--spawn--> worker (--agent-worker)
//
// Reusing this executable avoids packaging and versioning a separate launcher
// binary. The process boundary, rather than Go itself, provides the safety.
// PHP may be hosted by a multithreaded runtime such as FrankenPHP. Calling
// fork/daemon in the extension would leave the child with only one surviving
// thread but locks and runtime state inherited from threads that disappeared.
// posix_spawn loads the launcher as a clean process first, so detaching and
// starting the worker happen outside the PHP host.
//
// PHP waits for and reaps its short-lived launcher child. The launcher does not
// wait for the long-lived worker: when the launcher exits, the operating system
// reparents the worker to init or a subreaper.
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
	agentLockFD    = 3 // The first exec.Cmd.ExtraFiles entry becomes file descriptor 3.
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

// runLauncher is the normal entry point started by the PHP extension. Several
// PHP processes may start launchers concurrently, so each tries to lock the
// run directory. A launcher that cannot get the lock exits successfully because
// another worker is already running or being started. The winning launcher
// starts a detached worker, gives it the already-held lock, and then exits.
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
	// setsid separates the session and process group. The launcher's later exit,
	// not setsid, is what causes the worker to be reparented.
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	// The worker inherits the existing lock, so there is no unlocked handoff.
	command.ExtraFiles = []*os.File{agentLock}
	if err := command.Start(); err != nil {
		return err
	}
	// Do not wait for the long-lived worker. Returning exits the launcher, after
	// which init or a subreaper adopts the worker.
	return command.Process.Release()
}

// runWorker is the internal entry point started only by runLauncher with
// --agent-worker. It validates and retains the inherited run-directory lock,
// so the singleton lock is never released between launcher and worker. The
// worker owns the PID file and gRPC socket and serves all PHP processes until
// it is stopped.
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

// No arguments select the public launcher invocation. --agent-worker is a
// private marker used only when the launcher invokes the same executable again.
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
