package globals

import (
	"log"
	"os"
	"regexp"
	"sync"

	. "main/aikido_types"
)

// ===========================
// Server Configuration
// ===========================

var EnvironmentConfig EnvironmentConfigData
var Servers = make(map[string]*ServerData)
var ServersMutex sync.RWMutex

// ===========================
// Per-Thread Context Storage
// ===========================
// Thread-safe per-thread context storage for ZTS (Zend Thread Safety)
// Using pthread ID as key ensures each OS thread has isolated context

var (
	ContextInstances sync.Map // map[uint64]unsafe.Pointer - pthread ID -> instance pointer
	ContextData      sync.Map // map[uint64]*RequestContextData - pthread ID -> request context
	EventContextData sync.Map // map[uint64]*EventContextData - pthread ID -> event context
)

// ===========================
// Logging State
// ===========================

type LogLevel int

const (
	LogDebugLevel LogLevel = iota
	LogInfoLevel
	LogWarnLevel
	LogErrorLevel
)

var (
	CurrentLogLevel = LogErrorLevel
	Logger          = log.New(os.Stdout, "", 0)
	CliLogging      = true
	LogFilePath     = ""
	LogMutex        sync.RWMutex
	LogFile         *os.File
)

func NewServerData() *ServerData {
	return &ServerData{
		AikidoConfig: AikidoConfigData{},
		CloudConfig: CloudConfigData{
			Block: -1,
		},
		CloudConfigMutex:    sync.Mutex{},
		MiddlewareInstalled: false,
		ParamMatchers:       make(map[string]*regexp.Regexp),
	}
}

func GetServer(token string) *ServerData {
	if token == "" {
		return nil
	}
	ServersMutex.RLock()
	defer ServersMutex.RUnlock()
	return Servers[token]
}

func GetServers() []*ServerData {
	ServersMutex.RLock()
	defer ServersMutex.RUnlock()
	servers := []*ServerData{}
	for _, server := range Servers {
		servers = append(servers, server)
	}
	return servers
}

func ServerExists(token string) bool {
	ServersMutex.RLock()
	defer ServersMutex.RUnlock()
	_, exists := Servers[token]
	return exists
}

func CreateServer(token string) *ServerData {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()
	Servers[token] = NewServerData()
	return Servers[token]
}

const (
	Version    = "1.4.13"
	SocketPath = "/run/aikido-" + Version + "/aikido-agent.sock"
)

func GetFromThreadStorage[T any](threadID uint64, storage *sync.Map) T {
	if val, ok := storage.Load(threadID); ok {
		return val.(T)
	}
	var zero T
	return zero
}

func StoreInThreadStorage(threadID uint64, data interface{}, storage *sync.Map) {
	storage.Store(threadID, data)
}

func LoadOrStoreInThreadStorage[T any](threadID uint64, newData T, storage *sync.Map) T {
	if val, ok := storage.Load(threadID); ok {
		return val.(T)
	}
	storage.Store(threadID, newData)
	return newData
}

func DeleteFromThreadStorage(threadID uint64, storage *sync.Map) {
	storage.Delete(threadID)
}

func HasInThreadStorage(threadID uint64, storage *sync.Map) bool {
	_, ok := storage.Load(threadID)
	return ok
}
