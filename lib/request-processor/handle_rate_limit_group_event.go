package main

import (
	"main/context"
	"main/log"
)

func OnRateLimitGroupEvent() string {
	context.ContextSetRateLimitGroup()
	group := context.GetRateLimitGroup()
	log.Infof("Got rate limit group: %s", group)
	return ""
}
