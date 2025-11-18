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
	"time"
	"unsafe"
)

var eventHandlers = map[int]HandlerFunction{
	C.EVENT_PRE_REQUEST: func(i interface{}) string {
		return OnPreRequest(i.(*instance.RequestProcessorInstance))
	},
	C.EVENT_POST_REQUEST: func(i interface{}) string {
		return OnPostRequest(i.(*instance.RequestProcessorInstance))
	},
	C.EVENT_SET_USER: func(i interface{}) string {
		return OnUserEvent(i.(*instance.RequestProcessorInstance))
	},
	C.EVENT_SET_RATE_LIMIT_GROUP: func(i interface{}) string {
		return OnRateLimitGroupEvent(i.(*instance.RequestProcessorInstance))
	},
	C.EVENT_GET_AUTO_BLOCKING_STATUS: func(i interface{}) string {
		return OnGetAutoBlockingStatus(i.(*instance.RequestProcessorInstance))
	},
	C.EVENT_GET_BLOCKING_STATUS: func(i interface{}) string {
		return OnGetBlockingStatus(i.(*instance.RequestProcessorInstance))
	},
	C.EVENT_PRE_OUTGOING_REQUEST: func(i interface{}) string {
		return OnPreOutgoingRequest(i.(*instance.RequestProcessorInstance))
	},
	C.EVENT_POST_OUTGOING_REQUEST: func(i interface{}) string {
		return OnPostOutgoingRequest(i.(*instance.RequestProcessorInstance))
	},
	C.EVENT_PRE_SHELL_EXECUTED: func(i interface{}) string {
		return OnPreShellExecuted(i.(*instance.RequestProcessorInstance))
	},
	C.EVENT_PRE_PATH_ACCESSED: func(i interface{}) string {
		return OnPrePathAccessed(i.(*instance.RequestProcessorInstance))
	},
	C.EVENT_PRE_SQL_QUERY_EXECUTED: func(i interface{}) string {
		return OnPreSqlQueryExecuted(i.(*instance.RequestProcessorInstance))
	},
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
	defer func() {
		if r := recover(); r != nil {
			log.Warn("Recovered from panic:", r)
			initOk = false
		}
	}()

	inst := instance.GetInstance(instancePtr)
	if inst == nil {
		return false
	}

	config.Init(inst, initJson)

	log.Debugf("Aikido Request Processor v%s (server PID: %d, request processor PID: %d) started in \"%s\" mode!",
		globals.Version,
		globals.EnvironmentConfig.ServerPID,
		globals.EnvironmentConfig.RequestProcessorPID,
		globals.EnvironmentConfig.PlatformName,
	)
	log.Debugf("Init data: %s", initJson)
	log.Debugf("Started with token: \"AIK_RUNTIME_***%s\"", utils.AnonymizeToken(inst.GetCurrentToken()))

	if globals.EnvironmentConfig.PlatformName != "cli" {
		grpc.Init()
		server := inst.GetCurrentServer()
		if server != nil {
			initializeServer(server)
		}
		grpc.StartCloudConfigRoutine()
	}
	if !zen_internals.Init() {
		log.Error("Error initializing zen-internals library!")
		return false
	}
	return true
}

func GoContextCallback(contextId int) string {
	// Get the instance from the context package
	// This works because context.Init stores the instance pointer
	instPtr := context.GetInstancePtr()
	inst := instance.GetInstance(instPtr)
	if inst == nil {
		return ""
	}

	contextCallbackPtr := inst.GetContextCallback()
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
	defer func() {
		if r := recover(); r != nil {
			log.Warn("Recovered from panic:", r)
			initOk = false
		}
	}()

	inst := instance.GetInstance(instancePtr)
	if inst == nil {
		return false
	}

	inst.SetContextCallback(unsafe.Pointer(contextCallback))
	return context.Init(instancePtr, GoContextCallback)
}

/*
	RequestProcessorConfigUpdate is used to update the Aikido Config loaded from env variables and send this config via gRPC to the Aikido Agent.
*/
//export RequestProcessorConfigUpdate
func RequestProcessorConfigUpdate(instancePtr unsafe.Pointer, configJson string) (initOk bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Warn("Recovered from panic:", r)
			initOk = false
		}
	}()

	inst := instance.GetInstance(instancePtr)

	if inst == nil {
		return false
	}

	log.Debugf("Reloading Aikido config...")
	conf := AikidoConfigData{}

	reloadResult := config.ReloadAikidoConfig(inst, &conf, configJson)

	server := inst.GetCurrentServer()

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
	defer func() {
		if r := recover(); r != nil {
			log.Warn("Recovered from panic:", r)
			outputJson = nil
		}
	}()

	inst := instance.GetInstance(instancePtr)
	if inst == nil {
		return nil
	}

	handler, exists := eventHandlers[eventId]
	if !exists {
		return nil
	}

	goString := handler(inst)
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
	inst := instance.GetInstance(instancePtr)
	if inst == nil {
		return -1
	}
	return utils.GetBlockingMode(inst.GetCurrentServer())
}

//export RequestProcessorReportStats
func RequestProcessorReportStats(instancePtr unsafe.Pointer, sink, kind string, attacksDetected, attacksBlocked, interceptorThrewError, withoutContext, total int32, timings []int64) {
	if globals.EnvironmentConfig.PlatformName == "cli" {
		return
	}

	inst := instance.GetInstance(instancePtr)
	if inst == nil {
		return
	}

	clonedTimings := make([]int64, len(timings))
	copy(clonedTimings, timings)
	go grpc.OnMonitoredSinkStats(inst.GetCurrentServer(), strings.Clone(sink), strings.Clone(kind), attacksDetected, attacksBlocked, interceptorThrewError, withoutContext, total, clonedTimings)
}

//export RequestProcessorUninit
func RequestProcessorUninit(instancePtr unsafe.Pointer) {
	log.Debug("Uninit: {}")
	zen_internals.Uninit()

	if globals.EnvironmentConfig.PlatformName != "cli" {
		grpc.Uninit()
	}

	log.Debugf("Aikido Request Processor v%s stopped!", globals.Version)
	config.Uninit()
}

func main() {}
