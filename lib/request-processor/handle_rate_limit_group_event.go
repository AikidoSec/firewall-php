package main

import (
	"main/context"
	"main/instance"
	"main/log"
)

func OnRateLimitGroupEvent(instance *instance.RequestProcessorInstance) string {
	context.ContextSetRateLimitGroup(instance)
	group := context.GetRateLimitGroup(instance)
	log.Infof(instance, "Got rate limit group: %s", group)
	return ""
}
