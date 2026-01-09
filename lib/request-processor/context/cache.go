package context

// #include "../../API.h"
import "C"
import (
	"main/helpers"
	"main/instance"
	"main/log"
	"main/utils"
	"strconv"
)

/*
	Context caching functions are present in this package.
	These cache data for each request instance.
	In this way, the code can request data on demand from the request context cache,
	and if the data it not yet initialized, only than it is requested from PHP (C++ extension) via callback.
	This allows to copy data from PHP only once per request and only when needed.
*/

type ParseFunction func(string) map[string]interface{}

func ContextSetMap(inst *instance.RequestProcessorInstance, contextId int, rawDataPtr **string, parsedPtr **map[string]interface{}, stringsPtr **map[string]string, parseFunc ParseFunction) {
	if stringsPtr != nil && *stringsPtr != nil {
		return
	}

	c := GetContext(inst)
	if c.Callback == nil {
		return
	}

	contextData := c.Callback(inst, contextId)
	if rawDataPtr != nil {
		*rawDataPtr = &contextData
	}
	if parsedPtr != nil {
		parsed := parseFunc(contextData)
		*parsedPtr = &parsed
		if stringsPtr != nil {
			strings := helpers.ExtractStringsFromUserInput(parsed, []helpers.PathPart{}, 0)
			*stringsPtr = &strings
		}
	}
}

func ContextSetString(inst *instance.RequestProcessorInstance, context_id int, m **string) {
	if *m != nil {
		return
	}

	c := GetContext(inst)
	if c.Callback == nil {
		return
	}

	temp := c.Callback(inst, context_id)
	*m = &temp
}

func ContextSetBody(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	ContextSetMap(inst, C.CONTEXT_BODY, &c.BodyRaw, &c.BodyParsed, &c.BodyParsedFlattened, utils.ParseBody)
}

func ContextSetQuery(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	ContextSetMap(inst, C.CONTEXT_QUERY, nil, &c.QueryParsed, &c.QueryParsedFlattened, utils.ParseQuery)
}

func ContextSetCookies(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	ContextSetMap(inst, C.CONTEXT_COOKIES, nil, &c.CookiesParsed, &c.CookiesParsedFlattened, utils.ParseCookies)
}

func ContextSetHeaders(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	ContextSetMap(inst, C.CONTEXT_HEADERS, nil, &c.HeadersParsed, &c.HeadersParsedFlattened, utils.ParseHeaders)
}

func ContextSetRouteParams(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	ContextSetMap(inst, C.CONTEXT_ROUTE, &c.RouteParamsRaw, &c.RouteParamsParsed, &c.RouteParamsParsedFlattened, utils.ParseRouteParams)
}

func ContextSetStatusCode(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	if c.StatusCode != nil {
		return
	}
	if c.Callback == nil {
		return
	}
	status_code_str := c.Callback(inst, C.CONTEXT_STATUS_CODE)
	status_code, err := strconv.Atoi(status_code_str)
	if err != nil {
		log.Warnf(inst, "Error parsing status code %v: %v", status_code_str, err)
		return
	}
	c.StatusCode = &status_code
}

func ContextSetRoute(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	ContextSetString(inst, C.CONTEXT_ROUTE, &c.Route)
}

func ContextSetParsedRoute(inst *instance.RequestProcessorInstance) {
	parsedRoute := utils.BuildRouteFromURL(inst, GetRoute(inst))
	c := GetContext(inst)
	c.RouteParsed = &parsedRoute
}

func ContextSetMethod(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	ContextSetString(inst, C.CONTEXT_METHOD, &c.Method)
}

func ContextSetUrl(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	ContextSetString(inst, C.CONTEXT_URL, &c.URL)
}

func ContextSetUserAgent(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	ContextSetString(inst, C.CONTEXT_HEADER_USER_AGENT, &c.UserAgent)
}

func ContextSetIp(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	if c.IP != nil {
		return
	}
	if c.Callback == nil {
		return
	}
	remoteAddress := c.Callback(inst, C.CONTEXT_REMOTE_ADDRESS)
	xForwardedFor := c.Callback(inst, C.CONTEXT_HEADER_X_FORWARDED_FOR)

	server := c.inst.GetCurrentServer()
	ip := utils.GetIpFromRequest(server, remoteAddress, xForwardedFor)
	c.IP = &ip
}

func ContextSetIsIpBypassed(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	if c.IsIpBypassed != nil {
		return
	}

	server := c.inst.GetCurrentServer()
	isIpBypassed := utils.IsIpBypassed(inst, server, GetIp(inst))
	c.IsIpBypassed = &isIpBypassed
}

func ContextSetUserId(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	ContextSetString(inst, C.CONTEXT_USER_ID, &c.UserId)
}

func ContextSetUserName(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	ContextSetString(inst, C.CONTEXT_USER_NAME, &c.UserName)
}

func ContextSetRateLimitGroup(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	if c.RateLimitGroup != nil {
		return
	}
	if c.Callback == nil {
		return
	}
	rateLimitGroup := c.Callback(inst, C.CONTEXT_RATE_LIMIT_GROUP)
	c.RateLimitGroup = &rateLimitGroup
}

func ContextSetEndpointConfig(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	if c.EndpointConfig != nil {
		return
	}

	// Per-thread isolation via sync.Map prevents context bleeding
	server := c.inst.GetCurrentServer()
	if server == nil {
		return
	}

	method := GetMethod(inst)
	route := GetParsedRoute(inst)
	endpointConfig := utils.GetEndpointConfig(server, method, route)
	c.EndpointConfig = &endpointConfig
}

func ContextSetWildcardEndpointsConfigs(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	if c.WildcardEndpointsConfigs != nil {
		return
	}

	// Per-thread isolation via sync.Map prevents context bleeding
	server := c.inst.GetCurrentServer()
	if server == nil {
		return
	}

	wildcardEndpointsConfigs := utils.GetWildcardEndpointsConfigs(server, GetMethod(inst), GetParsedRoute(inst))
	c.WildcardEndpointsConfigs = &wildcardEndpointsConfigs
}

func ContextSetIsEndpointProtectionTurnedOff(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	if c.IsEndpointProtectionTurnedOff != nil {
		return
	}

	isEndpointProtectionTurnedOff := false

	endpointConfig := GetEndpointConfig(inst)
	if endpointConfig != nil {
		isEndpointProtectionTurnedOff = endpointConfig.ForceProtectionOff
	}
	if !isEndpointProtectionTurnedOff {
		for _, wildcardEndpointConfig := range GetWildcardEndpointsConfig(inst) {
			if wildcardEndpointConfig.ForceProtectionOff {
				isEndpointProtectionTurnedOff = true
				break
			}
		}
	}
	c.IsEndpointProtectionTurnedOff = &isEndpointProtectionTurnedOff
}

func ContextSetIsEndpointConfigured(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	if c.IsEndpointConfigured != nil {
		return
	}

	IsEndpointConfigured := false

	endpointConfig := GetEndpointConfig(inst)
	if endpointConfig != nil {
		IsEndpointConfigured = true
	}
	if !IsEndpointConfigured {
		if len(GetWildcardEndpointsConfig(inst)) != 0 {
			IsEndpointConfigured = true
		}
	}
	c.IsEndpointConfigured = &IsEndpointConfigured
}

func ContextSetIsEndpointRateLimitingEnabled(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	if c.IsEndpointRateLimitingEnabled != nil {
		return
	}

	IsEndpointRateLimitingEnabled := false

	endpointConfig := GetEndpointConfig(inst)
	if endpointConfig != nil {
		IsEndpointRateLimitingEnabled = endpointConfig.RateLimiting.Enabled
	}
	if !IsEndpointRateLimitingEnabled {
		for _, wildcardEndpointConfig := range GetWildcardEndpointsConfig(inst) {
			if wildcardEndpointConfig.RateLimiting.Enabled {
				IsEndpointRateLimitingEnabled = true
				break
			}
		}
	}
	c.IsEndpointRateLimitingEnabled = &IsEndpointRateLimitingEnabled
}

func ContextSetIsEndpointIpAllowed(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	if c.IsEndpointIpAllowed != nil {
		return
	}

	ip := GetIp(inst)

	isEndpointIpAllowed := utils.NoConfig

	server := c.inst.GetCurrentServer()
	endpointConfig := GetEndpointConfig(inst)
	if endpointConfig != nil && server != nil {
		isEndpointIpAllowed = utils.IsIpAllowedOnEndpoint(inst, server, endpointConfig.AllowedIPAddresses, ip)
	}

	if isEndpointIpAllowed == utils.NoConfig && server != nil {
		for _, wildcardEndpointConfig := range GetWildcardEndpointsConfig(inst) {
			isEndpointIpAllowed = utils.IsIpAllowedOnEndpoint(inst, server, wildcardEndpointConfig.AllowedIPAddresses, ip)
			if isEndpointIpAllowed != utils.NoConfig {
				break
			}
		}
	}

	isEndpointIpAllowedBool := isEndpointIpAllowed != utils.NotFound

	c.IsEndpointIpAllowed = &isEndpointIpAllowedBool
}

func ContextSetIsEndpointRateLimited(inst *instance.RequestProcessorInstance) {
	c := GetContext(inst)
	c.IsEndpointRateLimited = true
}
