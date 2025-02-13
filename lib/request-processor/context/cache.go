package context

// #include "../../API.h"
import "C"
import (
	"main/helpers"
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

func ContextSetMap(contextId int, rawDataPtr **string, parsedPtr **map[string]interface{}, stringsPtr **map[string]string, parseFunc ParseFunction) {
	if stringsPtr != nil && *stringsPtr != nil {
		return
	}

	contextData := Context.Callback(contextId)
	if rawDataPtr != nil {
		*rawDataPtr = &contextData
	}
	if parsedPtr != nil {
		parsed := parseFunc(contextData)
		*parsedPtr = &parsed
		if stringsPtr != nil {
			strings := helpers.ExtractStringsFromUserInput(parsed, []helpers.PathPart{})
			*stringsPtr = &strings
		}
	}
}

func ContextSetString(context_id int, m **string) {
	if *m != nil {
		return
	}
	temp := Context.Callback(context_id)
	*m = &temp
}

func ContextSetBody() {
	ContextSetMap(C.CONTEXT_BODY, &Context.BodyRaw, &Context.BodyParsed, &Context.BodyParsedFlattened, utils.ParseBody)
}

func ContextSetQuery() {
	ContextSetMap(C.CONTEXT_QUERY, nil, &Context.QueryParsed, &Context.QueryParsedFlattened, utils.ParseQuery)
}

func ContextSetCookies() {
	ContextSetMap(C.CONTEXT_COOKIES, nil, &Context.CookiesParsed, &Context.CookiesParsedFlattened, utils.ParseCookies)
}

func ContextSetHeaders() {
	ContextSetMap(C.CONTEXT_HEADERS, nil, &Context.HeadersParsed, &Context.HeadersParsedFlattened, utils.ParseHeaders)
}

func ContextSetStatusCode() {
	if Context.StatusCode != nil {
		return
	}
	status_code_str := Context.Callback(C.CONTEXT_STATUS_CODE)
	status_code, err := strconv.Atoi(status_code_str)
	if err != nil {
		log.Warnf("Error parsing status code %v: %v", status_code_str, err)
		return
	}
	Context.StatusCode = &status_code
}

func ContextSetRoute() {
	ContextSetString(C.CONTEXT_ROUTE, &Context.Route)
}

func ContextSetParsedRoute() {
	parsedRoute := utils.BuildRouteFromURL(GetRoute())
	Context.RouteParsed = &parsedRoute
}

func ContextSetMethod() {
	ContextSetString(C.CONTEXT_METHOD, &Context.Method)
}

func ContextSetUrl() {
	ContextSetString(C.CONTEXT_URL, &Context.URL)
}

func ContextSetUserAgent() {
	ContextSetString(C.CONTEXT_HEADER_USER_AGENT, &Context.UserAgent)
}

func ContextSetIp() {
	if Context.IP != nil {
		return
	}
	remoteAddress := Context.Callback(C.CONTEXT_REMOTE_ADDRESS)
	xForwardedFor := Context.Callback(C.CONTEXT_HEADER_X_FORWARDED_FOR)
	ip := utils.GetIpFromRequest(remoteAddress, xForwardedFor)
	Context.IP = &ip
}

func ContextSetIsIpBypassed() {
	if Context.IsIpBypassed != nil {
		return
	}

	isIpBypassed := utils.IsIpBypassed(GetIp())
	Context.IsIpBypassed = &isIpBypassed
}

func ContextSetUserId() {
	ContextSetString(C.CONTEXT_USER_ID, &Context.UserId)
}

func ContextSetUserName() {
	ContextSetString(C.CONTEXT_USER_NAME, &Context.UserName)
}

func ContextSetIsProtectionTurnedOff() {
	if Context.IsProtectionTurnedOff != nil {
		return
	}

	method := GetMethod()
	route := GetParsedRoute()

	endpointConfig := utils.GetEndpointConfig(method, route)
	isProtectionTurnedOff := false
	if endpointConfig != nil {
		isProtectionTurnedOff = endpointConfig.ForceProtectionOff
	}
	Context.IsProtectionTurnedOff = &isProtectionTurnedOff
}
