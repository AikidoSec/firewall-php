package main

import (
	"main/context"
	"main/globals"
	"main/grpc"
	"main/log"
)

func OnRateLimitGroupEvent() string {
	if !globals.MiddlewareInstalled {
		go grpc.OnMiddlewareInstalled()
		globals.MiddlewareInstalled = true
	}

	context.ContextSetRateLimitGroup()
	group := context.GetRateLimitGroup()
	log.Infof("Got rate limit group: %s", group)
	return ""
}
