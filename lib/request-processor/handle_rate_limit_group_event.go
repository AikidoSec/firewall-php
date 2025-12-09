package main

import (
	"main/context"
	"main/instance"
	"main/log"
)

func OnRateLimitGroupEvent(inst *instance.RequestProcessorInstance) string {
	context.ContextSetRateLimitGroup(inst)
	group := context.GetRateLimitGroup(inst)
	log.Infof(inst, "Got rate limit group: %s", group)
	return ""
}
