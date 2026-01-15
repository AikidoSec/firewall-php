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
	inst                          *instance.RequestProcessorInstance // CACHED: Instance pointer for fast access
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
	inst := instance.GetInstance(instPtr)
	if inst == nil {
		return false
	}

	inst.SetContextInstance(instPtr)

	ctx := &RequestContextData{
		inst:     inst,
		Callback: callback,
	}
	inst.SetRequestContext(ctx)

	inst.SetEventContext(&EventContextData{})

	return true
}

func GetContext(inst *instance.RequestProcessorInstance) *RequestContextData {
	if inst == nil {
		return nil
	}
	ctx := inst.GetRequestContext()
	if ctx == nil {
		return nil
	}
	return ctx.(*RequestContextData)
}

func (ctx *RequestContextData) GetInstance() *instance.RequestProcessorInstance {
	return ctx.inst
}

func Clear(inst *instance.RequestProcessorInstance) bool {
	ctx := GetContext(inst)
	*ctx = RequestContextData{
		inst:     inst,
		Callback: ctx.Callback,
	}
	ResetEventContext(inst)
	return true
}

func GetFromCache[T any](inst *instance.RequestProcessorInstance, fetchDataFn func(*instance.RequestProcessorInstance), s **T) T {
	if fetchDataFn != nil {
		fetchDataFn(inst)
	}
	if *s == nil {
		var t T
		c := GetContext(inst)
		if c != nil && c.inst != nil && inst.GetCurrentServer() != nil {
			log.Warnf(inst, "Error getting from cache. Returning default value %v...", t)
		}
		return t
	}
	return **s
}

func GetMethod(inst *instance.RequestProcessorInstance) string {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetMethod, &ctx.Method)
}

func GetRoute(inst *instance.RequestProcessorInstance) string {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetRoute, &ctx.Route)
}

func GetIp(inst *instance.RequestProcessorInstance) string {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetIp, &ctx.IP)
}

func GetUserId(inst *instance.RequestProcessorInstance) string {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetUserId, &ctx.UserId)
}

func GetUserAgent(inst *instance.RequestProcessorInstance) string {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetUserAgent, &ctx.UserAgent)
}

func GetStatusCode(inst *instance.RequestProcessorInstance) int {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetStatusCode, &ctx.StatusCode)
}

func GetUrl(inst *instance.RequestProcessorInstance) string {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetUrl, &ctx.URL)
}

func GetBodyRaw(inst *instance.RequestProcessorInstance) string {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetBody, &ctx.BodyRaw)
}

func GetParsedRoute(inst *instance.RequestProcessorInstance) string {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetParsedRoute, &ctx.RouteParsed)
}

func GetRateLimitGroup(inst *instance.RequestProcessorInstance) string {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetRateLimitGroup, &ctx.RateLimitGroup)
}

func GetQueryParsed(inst *instance.RequestProcessorInstance) map[string]interface{} {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetQuery, &ctx.QueryParsed)
}

func GetHeadersParsed(inst *instance.RequestProcessorInstance) map[string]interface{} {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetHeaders, &ctx.HeadersParsed)
}

func IsIpBypassed(inst *instance.RequestProcessorInstance) bool {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetIsIpBypassed, &ctx.IsIpBypassed)
}

func IsEndpointRateLimited(inst *instance.RequestProcessorInstance) bool {
	return GetContext(inst).IsEndpointRateLimited
}

func IsEndpointProtectionTurnedOff(inst *instance.RequestProcessorInstance) bool {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetIsEndpointProtectionTurnedOff, &ctx.IsEndpointProtectionTurnedOff)
}

func GetBodyParsed(inst *instance.RequestProcessorInstance) map[string]interface{} {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetBody, &ctx.BodyParsed)
}

func GetCookiesParsed(inst *instance.RequestProcessorInstance) map[string]interface{} {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetCookies, &ctx.CookiesParsed)
}

func GetBodyParsedFlattened(inst *instance.RequestProcessorInstance) map[string]string {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetBody, &ctx.BodyParsedFlattened)
}

func GetQueryParsedFlattened(inst *instance.RequestProcessorInstance) map[string]string {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetQuery, &ctx.QueryParsedFlattened)
}

func GetCookiesParsedFlattened(inst *instance.RequestProcessorInstance) map[string]string {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetCookies, &ctx.CookiesParsedFlattened)
}

func GetRouteParamsParsedFlattened(inst *instance.RequestProcessorInstance) map[string]string {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetRouteParams, &ctx.RouteParamsParsedFlattened)
}

func GetHeadersParsedFlattened(inst *instance.RequestProcessorInstance) map[string]string {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetHeaders, &ctx.HeadersParsedFlattened)
}

func GetUserName(inst *instance.RequestProcessorInstance) string {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetUserName, &ctx.UserName)
}

func GetEndpointConfig(inst *instance.RequestProcessorInstance) *EndpointData {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetEndpointConfig, &ctx.EndpointConfig)
}

func GetWildcardEndpointsConfig(inst *instance.RequestProcessorInstance) []EndpointData {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetWildcardEndpointsConfigs, &ctx.WildcardEndpointsConfigs)
}

func IsEndpointConfigured(inst *instance.RequestProcessorInstance) bool {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetIsEndpointConfigured, &ctx.IsEndpointConfigured)
}

func IsEndpointRateLimitingEnabled(inst *instance.RequestProcessorInstance) bool {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetIsEndpointRateLimitingEnabled, &ctx.IsEndpointRateLimitingEnabled)
}

func IsEndpointIpAllowed(inst *instance.RequestProcessorInstance) bool {
	ctx := GetContext(inst)
	return GetFromCache(inst, ContextSetIsEndpointIpAllowed, &ctx.IsEndpointIpAllowed)
}
