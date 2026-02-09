package context

// #include "../../API.h"
import "C"
import (
	. "main/aikido_types"
	"main/globals"
	"main/instance"
	"main/log"
	"unsafe"
)

type CallbackFunction func(*instance.RequestProcessorInstance, int) string

/* Request level context cache (changes on each PHP request) */
type RequestContextData struct {
	instance                      *instance.RequestProcessorInstance // CACHED: Instance pointer for fast access
	Callback                      CallbackFunction                   // callback to access data from the PHP layer (C++ extension) about the current request and current event
	Method                        *string
	Route                         *string
	RouteParsed                   *string
	URL                           *string
	StatusCode                    *int
	IP                            *string
	RateLimitGroup                *string
	EndpointConfig                **EndpointData
	WildcardEndpointsConfigs      *[]EndpointData
	IsIpBypassed                  *bool
	IsEndpointConfigured          *bool
	IsEndpointRateLimitingEnabled *bool
	IsEndpointProtectionTurnedOff *bool
	IsEndpointIpAllowed           *bool
	IsEndpointRateLimited         bool
	UserAgent                     *string
	UserId                        *string
	UserName                      *string
	BodyRaw                       *string
	BodyParsed                    *map[string]interface{}
	BodyParsedFlattened           *map[string]string
	QueryParsed                   *map[string]interface{}
	QueryParsedFlattened          *map[string]string
	CookiesParsed                 *map[string]interface{}
	CookiesParsedFlattened        *map[string]string
	HeadersParsed                 *map[string]interface{}
	HeadersParsedFlattened        *map[string]string
	RouteParamsRaw                *string
	RouteParamsParsed             *map[string]interface{}
	RouteParamsParsedFlattened    *map[string]string
}

func GetServerPID() int32 {
	return globals.EnvironmentConfig.ServerPID
}

func Init(instPtr unsafe.Pointer, callback CallbackFunction) bool {
	instance := instance.GetInstance(instPtr)
	if instance == nil {
		return false
	}

	instance.SetContextInstance(instPtr)

	ctx := &RequestContextData{
		instance: instance,
		Callback: callback,
	}
	instance.SetRequestContext(ctx)

	instance.SetEventContext(&EventContextData{})

	return true
}

func GetContext(instance *instance.RequestProcessorInstance) *RequestContextData {
	if instance == nil {
		return nil
	}
	ctx := instance.GetRequestContext()
	if ctx == nil {
		return nil
	}
	return ctx.(*RequestContextData)
}

func (ctx *RequestContextData) GetInstance() *instance.RequestProcessorInstance {
	return ctx.instance
}

func Clear(instance *instance.RequestProcessorInstance) bool {
	ctx := GetContext(instance)
	*ctx = RequestContextData{
		instance: instance,
		Callback: ctx.Callback,
	}
	ResetEventContext(instance)
	return true
}

func GetFromCache[T any](instance *instance.RequestProcessorInstance, fetchDataFn func(*instance.RequestProcessorInstance), s **T) T {
	if fetchDataFn != nil {
		fetchDataFn(instance)
	}
	if *s == nil {
		var t T
		c := GetContext(instance)
		if c != nil && c.instance != nil && instance.GetCurrentServer() != nil {
			log.Warnf(instance, "Error getting from cache. Returning default value %v...", t)
		}
		return t
	}
	return **s
}

func GetMethod(instance *instance.RequestProcessorInstance) string {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetMethod, &ctx.Method)
}

func GetRoute(instance *instance.RequestProcessorInstance) string {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetRoute, &ctx.Route)
}

func GetIp(instance *instance.RequestProcessorInstance) string {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetIp, &ctx.IP)
}

func GetUserId(instance *instance.RequestProcessorInstance) string {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetUserId, &ctx.UserId)
}

func GetUserAgent(instance *instance.RequestProcessorInstance) string {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetUserAgent, &ctx.UserAgent)
}

func GetStatusCode(instance *instance.RequestProcessorInstance) int {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetStatusCode, &ctx.StatusCode)
}

func GetUrl(instance *instance.RequestProcessorInstance) string {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetUrl, &ctx.URL)
}

func GetBodyRaw(instance *instance.RequestProcessorInstance) string {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetBody, &ctx.BodyRaw)
}

func GetParsedRoute(instance *instance.RequestProcessorInstance) string {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetParsedRoute, &ctx.RouteParsed)
}

func GetRateLimitGroup(instance *instance.RequestProcessorInstance) string {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetRateLimitGroup, &ctx.RateLimitGroup)
}

func GetQueryParsed(instance *instance.RequestProcessorInstance) map[string]interface{} {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetQuery, &ctx.QueryParsed)
}

func GetHeadersParsed(instance *instance.RequestProcessorInstance) map[string]interface{} {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetHeaders, &ctx.HeadersParsed)
}

func IsIpBypassed(instance *instance.RequestProcessorInstance) bool {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetIsIpBypassed, &ctx.IsIpBypassed)
}

func IsEndpointRateLimited(instance *instance.RequestProcessorInstance) bool {
	return GetContext(instance).IsEndpointRateLimited
}

func IsEndpointProtectionTurnedOff(instance *instance.RequestProcessorInstance) bool {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetIsEndpointProtectionTurnedOff, &ctx.IsEndpointProtectionTurnedOff)
}

func GetBodyParsed(instance *instance.RequestProcessorInstance) map[string]interface{} {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetBody, &ctx.BodyParsed)
}

func GetCookiesParsed(instance *instance.RequestProcessorInstance) map[string]interface{} {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetCookies, &ctx.CookiesParsed)
}

func GetBodyParsedFlattened(instance *instance.RequestProcessorInstance) map[string]string {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetBody, &ctx.BodyParsedFlattened)
}

func GetQueryParsedFlattened(instance *instance.RequestProcessorInstance) map[string]string {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetQuery, &ctx.QueryParsedFlattened)
}

func GetCookiesParsedFlattened(instance *instance.RequestProcessorInstance) map[string]string {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetCookies, &ctx.CookiesParsedFlattened)
}

func GetRouteParamsParsedFlattened(instance *instance.RequestProcessorInstance) map[string]string {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetRouteParams, &ctx.RouteParamsParsedFlattened)
}

func GetHeadersParsedFlattened(instance *instance.RequestProcessorInstance) map[string]string {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetHeaders, &ctx.HeadersParsedFlattened)
}

func GetUserName(instance *instance.RequestProcessorInstance) string {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetUserName, &ctx.UserName)
}

func GetEndpointConfig(instance *instance.RequestProcessorInstance) *EndpointData {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetEndpointConfig, &ctx.EndpointConfig)
}

func GetWildcardEndpointsConfig(instance *instance.RequestProcessorInstance) []EndpointData {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetWildcardEndpointsConfigs, &ctx.WildcardEndpointsConfigs)
}

func IsEndpointConfigured(instance *instance.RequestProcessorInstance) bool {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetIsEndpointConfigured, &ctx.IsEndpointConfigured)
}

func IsEndpointRateLimitingEnabled(instance *instance.RequestProcessorInstance) bool {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetIsEndpointRateLimitingEnabled, &ctx.IsEndpointRateLimitingEnabled)
}

func IsEndpointIpAllowed(instance *instance.RequestProcessorInstance) bool {
	ctx := GetContext(instance)
	return GetFromCache(instance, ContextSetIsEndpointIpAllowed, &ctx.IsEndpointIpAllowed)
}
