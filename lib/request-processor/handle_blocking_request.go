package main

import (
	"encoding/json"
	"fmt"
	"html"
	"main/context"
	"main/globals"
	"main/grpc"
	"main/log"
	"main/utils"
	"time"
)

func GetAction(actionHandling, actionType, trigger, description, data string, responseCode int) string {
	actionMap := map[string]interface{}{
		"action":        actionHandling,
		"type":          actionType,
		"trigger":       trigger,
		"description":   html.EscapeString(description),
		"message":       fmt.Sprintf("Your %s (%s) is blocked due to: %s!", trigger, data, description),
		"data":          data,
		"response_code": responseCode,
	}
	actionJson, err := json.Marshal(actionMap)
	if err != nil {
		return ""
	}
	return string(actionJson)
}

func OnGetBlockingStatus() string {
	log.Debugf("OnGetBlockingStatus called!")

	if !globals.MiddlewareInstalled {
		go grpc.OnMiddlewareInstalled()
		globals.MiddlewareInstalled = true
	}

	userId := context.GetUserId()
	if utils.IsUserBlocked(userId) {
		log.Infof("User \"%s\" is blocked!", userId)
		return GetAction("store", "blocked", "user", "user blocked from config", userId, 403)
	}

	return OnGetAutoBlockingStatus()
}

func OnGetAutoBlockingStatus() string {
	log.Debugf("OnGetAutoBlockingStatus called!")

	method := context.GetMethod()
	route := context.GetParsedRoute()
	if method == "" || route == "" {
		return ""
	}

	ip := context.GetIp()
	userAgent := context.GetUserAgent()
	endpointData := utils.GetEndpointConfig(method, route)

	if endpointData != nil && !utils.IsIpAllowed(endpointData.AllowedIPAddresses, ip) {
		log.Infof("IP \"%s\" is not allowd to access this endpoint!", ip)
		return GetAction("exit", "blocked", "ip", "not allowed by config to access this endpoint", ip, 403)
	}

	if context.IsIpBypassed() {
		log.Infof("IP \"%s\" is bypassed! Skipping additional checks...", ip)
		return ""
	}

	if ipBlocked, ipBlockedDescription := utils.IsIpBlocked(ip); ipBlocked {
		log.Infof("IP \"%s\" blocked due to: %s!", ip, ipBlockedDescription)
		return GetAction("exit", "blocked", "ip", ipBlockedDescription, ip, 403)
	}

	if userAgentBlocked, userAgentBlockedDescription := utils.IsUserAgentBlocked(userAgent); userAgentBlocked {
		log.Infof("User Agent \"%s\" blocked due to: %s!", userAgent, userAgentBlockedDescription)
		return GetAction("exit", "blocked", "user-agent", userAgentBlockedDescription, userAgent, 429)
	}

	userId := context.GetUserId()
	if endpointData != nil && endpointData.RateLimiting.Enabled {
		// If request is monitored for rate limiting,
		// do a sync call via gRPC to see if the request should be blocked or not
		rateLimitingStatus := grpc.GetRateLimitingStatus(method, route, userId, ip, 10*time.Millisecond)
		if rateLimitingStatus != nil && rateLimitingStatus.Block {
			log.Infof("Request made from IP \"%s\" is ratelimited by \"%s\"!", ip, rateLimitingStatus.Trigger)
			return GetAction("exit", "ratelimited", rateLimitingStatus.Trigger, "configured rate limit exceeded by current ip", ip, 429)
		}
	}

	return ""
}
