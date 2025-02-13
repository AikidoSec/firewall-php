package main

import (
	"encoding/json"
	"html"
	"main/context"
	"main/grpc"
	"main/log"
	"main/utils"
	"time"
)

func GetStoreAction(actionType, trigger, description, ip string) string {
	actionMap := map[string]interface{}{
		"action":      "store",
		"type":        actionType,
		"trigger":     trigger,
		"description": html.EscapeString(description),
	}
	if trigger == "ip" {
		actionMap["ip"] = ip
	}
	actionJson, err := json.Marshal(actionMap)
	if err != nil {
		return ""
	}
	return string(actionJson)
}

func OnGetBlockingStatus() string {
	go grpc.OnMiddlewareInstalled()

	userId := context.GetUserId()
	if utils.IsUserBlocked(userId) {
		log.Infof("User \"%s\" is blocked!", userId)
		return GetStoreAction("blocked", "user", "user blocked from config", "")
	}

	method := context.GetMethod()
	route := context.GetParsedRoute()
	if method == "" || route == "" {
		return ""
	}

	ip := context.GetIp()
	endpointData := utils.GetEndpointConfig(method, route)

	if endpointData != nil && !utils.IsIpAllowed(endpointData.AllowedIPAddresses, ip) {
		log.Infof("IP \"%s\" is not allowd to access this endpoint!", ip)
		return GetStoreAction("blocked", "ip", "not allowed by config to access this endpoint", ip)
	}

	if context.IsIpBypassed() {
		log.Infof("IP \"%s\" is bypassed! Skipping additional checks...", ip)
		return ""
	}

	if ipBlocked, ipBlockedDescription := utils.IsIpBlocked(ip); ipBlocked {
		log.Infof("IP \"%s\" blocked due to: %s!", ip, ipBlockedDescription)
		return GetStoreAction("blocked", "ip", ipBlockedDescription, ip)
	}

	if endpointData != nil && endpointData.RateLimiting.Enabled {
		// If request is monitored for rate limiting,
		// do a sync call via gRPC to see if the request should be blocked or not
		rateLimitingStatus := grpc.GetRateLimitingStatus(method, route, userId, ip, 10*time.Millisecond)
		if rateLimitingStatus != nil && rateLimitingStatus.Block {
			log.Infof("Request made from IP \"%s\" is ratelimited by \"%s\"!", ip, rateLimitingStatus.Trigger)
			return GetStoreAction("ratelimited", rateLimitingStatus.Trigger, "configured rate limit exceeded by current ip", ip)
		}
	}

	return ""
}
