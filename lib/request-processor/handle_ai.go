package main

// #include "../API.h"
import "C"
import (
	"encoding/json"
	"main/context"
	"main/grpc"
	"main/instance"
	"main/ipc/protos"
)

func OnAICall(instance *instance.RequestProcessorInstance) string {
	if instance.GetCurrentServer() == nil {
		return ""
	}
	call := &protos.AIUsage{}
	raw := context.GetContext(instance).Callback(instance, C.AI_CALL_STATS)
	if json.Unmarshal([]byte(raw), call) == nil && call.Provider != "" && call.Model != "" {
		go grpc.OnAIUsage(instance.GetCurrentToken(), []*protos.AIUsage{call})
	}
	return ""
}
