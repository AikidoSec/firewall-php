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

func ContextSetMap(instance *instance.RequestProcessorInstance, contextId int, rawDataPtr **string, parsedPtr **map[string]interface{}, stringsPtr **map[string]string, parseFunc ParseFunction) {
	if stringsPtr != nil && *stringsPtr != nil {
		return
	}

	c := GetContext(instance)
	if c.Callback == nil {
		return
	}

	contextData := c.Callback(instance, contextId)
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

func ContextSetString(instance *instance.RequestProcessorInstance, context_id int, m **string) {
	if *m != nil {
		return
	}

	c := GetContext(instance)
	if c.Callback == nil {
		return
	}

	temp := c.Callback(instance, context_id)
	*m = &temp
}

func ContextSetBody(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	ContextSetMap(instance, C.CONTEXT_BODY, &c.BodyRaw, &c.BodyParsed, &c.BodyParsedFlattened, utils.ParseBody)
}

func ContextSetQuery(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	ContextSetMap(instance, C.CONTEXT_QUERY, nil, &c.QueryParsed, &c.QueryParsedFlattened, utils.ParseQuery)
}

func ContextSetCookies(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	ContextSetMap(instance, C.CONTEXT_COOKIES, nil, &c.CookiesParsed, &c.CookiesParsedFlattened, utils.ParseCookies)
}

func ContextSetHeaders(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	ContextSetMap(instance, C.CONTEXT_HEADERS, nil, &c.HeadersParsed, &c.HeadersParsedFlattened, utils.ParseHeaders)
}

func ContextSetRouteParams(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	ContextSetMap(instance, C.CONTEXT_ROUTE, &c.RouteParamsRaw, &c.RouteParamsParsed, &c.RouteParamsParsedFlattened, utils.ParseRouteParams)
}

func ContextSetStatusCode(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	if c.StatusCode != nil {
		return
	}
	if c.Callback == nil {
		return
	}
	status_code_str := c.Callback(instance, C.CONTEXT_STATUS_CODE)
	status_code, err := strconv.Atoi(status_code_str)
	if err != nil {
		log.Warnf(instance, "Error parsing status code %v: %v", status_code_str, err)
		return
	}
	c.StatusCode = &status_code
}

func ContextSetRoute(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	ContextSetString(instance, C.CONTEXT_ROUTE, &c.Route)
}

func ContextSetParsedRoute(instance *instance.RequestProcessorInstance) {
	parsedRoute := utils.BuildRouteFromURL(instance, GetRoute(instance))
	c := GetContext(instance)
	c.RouteParsed = &parsedRoute
}

func ContextSetMethod(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	ContextSetString(instance, C.CONTEXT_METHOD, &c.Method)
}

func ContextSetUrl(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	ContextSetString(instance, C.CONTEXT_URL, &c.URL)
}

func ContextSetUserAgent(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	ContextSetString(instance, C.CONTEXT_HEADER_USER_AGENT, &c.UserAgent)
}

func ContextSetIp(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	if c.IP != nil {
		return
	}
	if c.Callback == nil {
		return
	}
	remoteAddress := c.Callback(instance, C.CONTEXT_REMOTE_ADDRESS)
	xForwardedFor := c.Callback(instance, C.CONTEXT_HEADER_X_FORWARDED_FOR)

	server := c.instance.GetCurrentServer()
	ip := utils.GetIpFromRequest(server, remoteAddress, xForwardedFor)
	c.IP = &ip
}

func ContextSetIsIpBypassed(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	if c.IsIpBypassed != nil {
		return
	}

	server := c.instance.GetCurrentServer()
	isIpBypassed := utils.IsIpBypassed(instance, server, GetIp(instance))
	c.IsIpBypassed = &isIpBypassed
}

func ContextSetUserId(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	ContextSetString(instance, C.CONTEXT_USER_ID, &c.UserId)
}

func ContextSetUserName(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	ContextSetString(instance, C.CONTEXT_USER_NAME, &c.UserName)
}

func ContextSetRateLimitGroup(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	if c.RateLimitGroup != nil {
		return
	}
	if c.Callback == nil {
		return
	}
	rateLimitGroup := c.Callback(instance, C.CONTEXT_RATE_LIMIT_GROUP)
	c.RateLimitGroup = &rateLimitGroup
}

func ContextSetEndpointConfig(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	if c.EndpointConfig != nil {
		return
	}

	server := c.instance.GetCurrentServer()
	if server == nil {
		return
	}

	method := GetMethod(instance)
	route := GetParsedRoute(instance)
	endpointConfig := utils.GetEndpointConfig(server, method, route)
	c.EndpointConfig = &endpointConfig
}

func ContextSetWildcardEndpointsConfigs(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	if c.WildcardEndpointsConfigs != nil {
		return
	}

	// Per-thread isolation via sync.Map prevents context bleeding
	server := c.instance.GetCurrentServer()
	if server == nil {
		return
	}

	wildcardEndpointsConfigs := utils.GetWildcardEndpointsConfigs(server, GetMethod(instance), GetParsedRoute(instance))
	c.WildcardEndpointsConfigs = &wildcardEndpointsConfigs
}

func ContextSetIsEndpointProtectionTurnedOff(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	if c.IsEndpointProtectionTurnedOff != nil {
		return
	}

	isEndpointProtectionTurnedOff := false

	endpointConfig := GetEndpointConfig(instance)
	if endpointConfig != nil {
		isEndpointProtectionTurnedOff = endpointConfig.ForceProtectionOff
	}
	if !isEndpointProtectionTurnedOff {
		for _, wildcardEndpointConfig := range GetWildcardEndpointsConfig(instance) {
			if wildcardEndpointConfig.ForceProtectionOff {
				isEndpointProtectionTurnedOff = true
				break
			}
		}
	}
	c.IsEndpointProtectionTurnedOff = &isEndpointProtectionTurnedOff
}

func ContextSetIsEndpointConfigured(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	if c.IsEndpointConfigured != nil {
		return
	}

	IsEndpointConfigured := false

	endpointConfig := GetEndpointConfig(instance)
	if endpointConfig != nil {
		IsEndpointConfigured = true
	}
	if !IsEndpointConfigured {
		if len(GetWildcardEndpointsConfig(instance)) != 0 {
			IsEndpointConfigured = true
		}
	}
	c.IsEndpointConfigured = &IsEndpointConfigured
}

func ContextSetIsEndpointRateLimitingEnabled(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	if c.IsEndpointRateLimitingEnabled != nil {
		return
	}

	IsEndpointRateLimitingEnabled := false

	endpointConfig := GetEndpointConfig(instance)
	if endpointConfig != nil {
		IsEndpointRateLimitingEnabled = endpointConfig.RateLimiting.Enabled
	}
	if !IsEndpointRateLimitingEnabled {
		for _, wildcardEndpointConfig := range GetWildcardEndpointsConfig(instance) {
			if wildcardEndpointConfig.RateLimiting.Enabled {
				IsEndpointRateLimitingEnabled = true
				break
			}
		}
	}
	c.IsEndpointRateLimitingEnabled = &IsEndpointRateLimitingEnabled
}

func ContextSetIsEndpointIpAllowed(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	if c.IsEndpointIpAllowed != nil {
		return
	}

	ip := GetIp(instance)

	isEndpointIpAllowed := utils.NoConfig

	server := c.instance.GetCurrentServer()
	endpointConfig := GetEndpointConfig(instance)
	if endpointConfig != nil && server != nil {
		isEndpointIpAllowed = utils.IsIpAllowedOnEndpoint(instance, server, endpointConfig.AllowedIPAddresses, ip)
	}

	if isEndpointIpAllowed == utils.NoConfig && server != nil {
		for _, wildcardEndpointConfig := range GetWildcardEndpointsConfig(instance) {
			isEndpointIpAllowed = utils.IsIpAllowedOnEndpoint(instance, server, wildcardEndpointConfig.AllowedIPAddresses, ip)
			if isEndpointIpAllowed != utils.NoConfig {
				break
			}
		}
	}

	isEndpointIpAllowedBool := isEndpointIpAllowed != utils.NotFound

	c.IsEndpointIpAllowed = &isEndpointIpAllowedBool
}

func ContextSetIsEndpointRateLimited(instance *instance.RequestProcessorInstance) {
	c := GetContext(instance)
	c.IsEndpointRateLimited = true
}
