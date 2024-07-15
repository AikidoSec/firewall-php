package main

import "C"
import (
	"encoding/json"
	"fmt"
	. "main/aikido_types"
	"main/globals"
	"main/grpc"
	"main/log"
	"main/utils"
)

var eventHandlers = map[string]HandlerFunction{
	"function_executed": OnFunctionExecuted,
	"method_executed":   OnMethodExecuted,
}

//export RequestProcessorOnEvent
func RequestProcessorOnEvent(eventJson string) (outputJson string) {
	defer func() {
		if r := recover(); r != nil {
			log.Warn("Recovered from panic:", r)
			outputJson = "{}"
		}
	}()

	log.Debug("OnEvent: ", eventJson)

	var event map[string]interface{}
	err := json.Unmarshal([]byte(eventJson), &event)
	if err != nil {
		panic(fmt.Sprintf("Error parsing JSON: %s", err))
	}

	eventName := utils.MustGetFromMap[string](event, "event")
	data := utils.MustGetFromMap[map[string]interface{}](event, "data")

	utils.CheckIfKeyExists(eventHandlers, eventName)

	return eventHandlers[eventName](data)
}

//export RequestProcessorInit
func RequestProcessorInit(initJson string) (initOk bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Warn("Recovered from panic:", r)
			initOk = false
		}
	}()

	err := json.Unmarshal([]byte(initJson), &globals.InitData)
	if err != nil {
		panic(fmt.Sprintf("Error parsing JSON: %s", err))
	}

	log.Init(globals.InitData.LogLevel)
	log.Debugf("Aikido Request Processor v%s started!", globals.Version)

	grpc.Init()

	return true
}

//export RequestProcessorUninit
func RequestProcessorUninit() {
	log.Debug("Uninit: {}")
	log.Debugf("Aikido Request Processor v%s stopped!", globals.Version)
}

func main() {}