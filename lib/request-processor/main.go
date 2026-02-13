package main

//#include "../API.h"
import "C"
import (
	. "main/aikido_types"
	"main/config"
	"main/context"
	"main/globals"
	"main/grpc"
	"main/instance"
	"main/log"
	"main/utils"
	zen_internals "main/vulnerabilities/zen-internals"
	"strings"
	"sync"
	"time"
	"unsafe"
)

type HandlerFunction func(*instance.RequestProcessorInstance) string

var grpcInitOnce sync.Once
var zenInternalsInitOnce sync.Once

var eventHandlers = map[int]HandlerFunction{
	C.EVENT_PRE_REQUEST:              OnPreRequest,
	C.EVENT_POST_REQUEST:             OnPostRequest,
	C.EVENT_SET_USER:                 OnUserEvent,
	C.EVENT_SET_RATE_LIMIT_GROUP:     OnRateLimitGroupEvent,
	C.EVENT_REGISTER_PARAM_MATCHER:   OnRegisterParamMatcherEvent,
	C.EVENT_GET_AUTO_BLOCKING_STATUS: OnGetAutoBlockingStatus,
	C.EVENT_GET_BLOCKING_STATUS:      OnGetBlockingStatus,
	C.EVENT_GET_IS_IP_BYPASSED:       OnGetIsIpBypassed,
	C.EVENT_PRE_OUTGOING_REQUEST:     OnPreOutgoingRequest,
	C.EVENT_POST_OUTGOING_REQUEST:    OnPostOutgoingRequest,
	C.EVENT_PRE_SHELL_EXECUTED:       OnPreShellExecuted,
	C.EVENT_PRE_PATH_ACCESSED:        OnPrePathAccessed,
	C.EVENT_PRE_SQL_QUERY_EXECUTED:   OnPreSqlQueryExecuted,
}

func initializeServer(server *ServerData) {
	grpc.SendAikidoConfig(server)
	grpc.OnPackages(server, server.AikidoConfig.Packages)
	grpc.GetCloudConfig(server, 5*time.Second)
}

//export CreateInstance
func CreateInstance(threadID uint64, isZTS bool) unsafe.Pointer {
	return instance.CreateInstance(threadID, isZTS)
}

//export DestroyInstance
func DestroyInstance(threadID uint64) {
	instance.DestroyInstance(threadID)
}

//export RequestProcessorInit
func RequestProcessorInit(instancePtr unsafe.Pointer, initJson string) (initOk bool) {
	instance := instance.GetInstance(instancePtr)
	defer func() {
		if r := recover(); r != nil {
			log.Warn(instance, "Recovered from panic:", r)
			initOk = false
		}
	}()

	if instance == nil {
		return false
	}

	config.Init(instance, initJson)

	log.Debugf(instance, "Aikido Request Processor v%s (server PID: %d, request processor PID: %d) started in \"%s\" mode!",
		globals.Version,
		globals.EnvironmentConfig.ServerPID,
		globals.EnvironmentConfig.RequestProcessorPID,
		globals.EnvironmentConfig.PlatformName,
	)
	log.Debugf(instance, "Init data: %s", initJson)
	log.Debugf(instance, "Started with token: \"AIK_RUNTIME_***%s\"", utils.AnonymizeToken(instance.GetCurrentToken()))

	if globals.EnvironmentConfig.PlatformName != "cli" {
		server := instance.GetCurrentServer()

		if server != nil {
			server.ServerInitMutex.Lock()
			defer server.ServerInitMutex.Unlock()

			if server.ServerInitialized {
				server.ServerInitMutex.Unlock()
				return true
			}
			initializeServer(server)
			server.ServerInitialized = true

		}
		grpcInitOnce.Do(func() {
			grpc.Init()
			grpc.StartCloudConfigRoutine()
		})
	}

	var zenInternalsInitSuccess = true
	zenInternalsInitOnce.Do(func() {
		zenInternalsInitSuccess = zen_internals.Init()
	})
	if !zenInternalsInitSuccess {
		log.Error(instance, "Error initializing zen-internals library!")
		return false
	}
	return true
}

func GoContextCallback(instance *instance.RequestProcessorInstance, contextId int) string {
	if instance == nil {
		return ""
	}

	contextCallbackPtr := instance.GetContextCallback()
	if contextCallbackPtr == nil {
		return ""
	}

	contextCallback := (C.ContextCallback)(contextCallbackPtr)
	contextData := C.call(contextCallback, C.int(contextId))
	if contextData == nil {
		return ""
	}

	goContextData := C.GoString(contextData)

	/*
		In order to pass dynamic strings from the PHP extension (C++), we need a dynamically allocated buffer, that is allocated by the C++ extension.
		This buffer needs to be freed by the RequestProcessor (Go) once it has finished copying the data.
	*/
	C.free(unsafe.Pointer(contextData))
	// Remove invalid UTF8 characters (normalize)
	goContextData = strings.ToValidUTF8(goContextData, "")
	return goContextData
}

//export RequestProcessorContextInit
func RequestProcessorContextInit(instancePtr unsafe.Pointer, contextCallback C.ContextCallback) (initOk bool) {
	instance := instance.GetInstance(instancePtr)
	defer func() {
		if r := recover(); r != nil {
			log.Warn(instance, "Recovered from panic:", r)
			initOk = false
		}
	}()

	if instance == nil {
		return false
	}

	instance.SetContextCallback(unsafe.Pointer(contextCallback))
	return context.Init(instancePtr, GoContextCallback)
}

/*
	RequestProcessorConfigUpdate is used to update the Aikido Config loaded from env variables and send this config via gRPC to the Aikido Agent.
*/
//export RequestProcessorConfigUpdate
func RequestProcessorConfigUpdate(instancePtr unsafe.Pointer, configJson string) (initOk bool) {
	instance := instance.GetInstance(instancePtr)
	defer func() {
		if r := recover(); r != nil {
			log.Warn(instance, "Recovered from panic:", r)
			initOk = false
		}
	}()

	if instance == nil {
		return false
	}

	log.Debugf(instance, "Reloading Aikido config...")
	conf := AikidoConfigData{}

	reloadResult := config.ReloadAikidoConfig(instance, &conf, configJson)

	server := instance.GetCurrentServer()

	if server == nil {
		return false
	}
	switch reloadResult {
	case config.ReloadWithNewToken:
		initializeServer(server)
		return true
	case config.ReloadWithPastSeenToken:
		grpc.GetCloudConfig(server, 5*time.Second)
		return true
	case config.ReloadWithSameToken:
		return true
	case config.ReloadError:
		return false
	}
	return false
}

//export RequestProcessorOnEvent
func RequestProcessorOnEvent(instancePtr unsafe.Pointer, eventId int) (outputJson *C.char) {
	instance := instance.GetInstance(instancePtr)
	defer func() {
		if r := recover(); r != nil {
			log.Warn(instance, "Recovered from panic:", r)
			outputJson = nil
		}
	}()

	if instance == nil {
		return nil
	}

	handler, exists := eventHandlers[eventId]
	if !exists {
		return nil
	}

	goString := handler(instance)
	if goString == "" {
		return nil
	}
	return C.CString(goString)
}

/*
	Returns -1 if the config was not yet pulled from Agent.
	Otherwise, if blocking was set from cloud, it returns that value.
	Otherwise, it returns the environment value.
*/
//export RequestProcessorGetBlockingMode
func RequestProcessorGetBlockingMode(instancePtr unsafe.Pointer) int {
	instance := instance.GetInstance(instancePtr)
	if instance == nil {
		return -1
	}
	return utils.GetBlockingMode(instance.GetCurrentServer())
}

//export RequestProcessorReportStats
func RequestProcessorReportStats(instancePtr unsafe.Pointer, sink, kind string, attacksDetected, attacksBlocked, interceptorThrewError, withoutContext, total int32, timings []int64) {
	if globals.EnvironmentConfig.PlatformName == "cli" {
		return
	}

	instance := instance.GetInstance(instancePtr)
	if instance == nil {
		return
	}

	server := instance.GetCurrentServer()
	if server == nil {
		return
	}

	clonedTimings := make([]int64, len(timings))
	copy(clonedTimings, timings)
	go grpc.OnMonitoredSinkStats(server, instance.GetCurrentToken(), strings.Clone(sink), strings.Clone(kind), attacksDetected, attacksBlocked, interceptorThrewError, withoutContext, total, clonedTimings)
}

//export RequestProcessorUninit
func RequestProcessorUninit(instancePtr unsafe.Pointer) {
	instance := instance.GetInstance(instancePtr)
	log.Debug(instance, "Uninit: {}")
	zen_internals.Uninit()

	if globals.EnvironmentConfig.PlatformName != "cli" {
		grpc.Uninit()
	}

	log.Debugf(instance, "Aikido Request Processor v%s stopped!", globals.Version)
	config.Uninit()
}

func main() {}
