package main

import (
	"main/context"
	"main/instance"
	"main/log"
)

func OnRateLimitGroupEvent(inst *instance.RequestProcessorInstance) string {
	context.ContextSetRateLimitGroup()
	group := context.GetRateLimitGroup()
	log.Infof("Got rate limit group: %s", group)
	return ""
}
