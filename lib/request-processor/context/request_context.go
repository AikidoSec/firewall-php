package context

// #include "../../API.h"
import "C"
import (
	"main/log"
)

type CallbackFunction func(int) string

/* Request level context cache (changes on each PHP request) */
type RequestContextData struct {
	Callback               CallbackFunction // callback to access data from the PHP layer (C++ extension) about the current request and current event
	Method                 *string
	Route                  *string
	RouteParsed            *string
	URL                    *string
	StatusCode             *int
	IP                     *string
	IsIpBypassed           *bool
	UserAgent              *string
	UserId                 *string
	UserName               *string
	BodyRaw                *string
	BodyParsedFlattened    *map[string]string
	QueryRaw               *string
	QueryParsedFlattened   *map[string]string
	Cookies                *string
	CookiesParsedFlattened *map[string]string
	Headers                *map[string]interface{}
	HeadersParsedFlattened *map[string]string
}

var Context RequestContextData

func Init(callback CallbackFunction) bool {
	if callback == nil {
		callback = Context.Callback
	}
	Context = RequestContextData{
		Callback: callback,
	}
	return true
}

func GetFromCache[T any](fetchDataFn func(), s **T) T {
	if fetchDataFn != nil {
		fetchDataFn()
	}
	if *s == nil {
		var t T
		log.Warnf("Error getting from cache. Returning default value %v...", t)
		return t
	}
	return **s
}

func GetIp() string {
	return GetFromCache(ContextSetIp, &Context.IP)
}

func GetMethod() string {
	return GetFromCache(ContextSetMethod, &Context.Method)
}

func GetRoute() string {
	return GetFromCache(ContextSetRoute, &Context.Route)
}

func GetParsedRoute() string {
	return GetFromCache(ContextSetParsedRoute, &Context.RouteParsed)
}

func GetUrl() string {
	return GetFromCache(ContextSetUrl, &Context.URL)
}

func GetStatusCode() int {
	return GetFromCache(ContextSetStatusCode, &Context.StatusCode)
}

func IsIpBypassed() bool {
	return GetFromCache(ContextSetIsIpBypassed, &Context.IsIpBypassed)
}

func GetBodyRaw() string {
	return GetFromCache(ContextSetBody, &Context.BodyRaw)
}

func GetBodyParsed() map[string]interface{} {
	return GetFromCache(ContextSetBody, &Context.BodyParsed)
}

func GetQueryParsed() map[string]interface{} {
	return GetFromCache(ContextSetQuery, &Context.QueryParsed)
}

func GetCookiesParsed() map[string]interface{} {
	return GetFromCache(ContextSetCookies, &Context.CookiesParsed)
}

func GetHeadersParsed() map[string]interface{} {
	return GetFromCache(ContextSetHeaders, &Context.HeadersParsed)
}

func GetBodyParsedFlattened() map[string]string {
	return GetFromCache(ContextSetBody, &Context.BodyParsedFlattened)
}

func GetQueryParsedFlattened() map[string]string {
	return GetFromCache(ContextSetQuery, &Context.QueryParsedFlattened)
}

func GetCookiesParsedFlattened() map[string]string {
	return GetFromCache(ContextSetCookies, &Context.CookiesParsedFlattened)
}

func GetHeadersParsedFlattened() map[string]string {
	return GetFromCache(ContextSetHeaders, &Context.HeadersParsedFlattened)
}

func GetUserAgent() string {
	return GetFromCache(ContextSetUserAgent, &Context.UserAgent)
}

func GetUserId() string {
	return GetFromCache(ContextSetUserId, &Context.UserId)
}

func GetUserName() string {
	return GetFromCache(ContextSetUserName, &Context.UserName)
}
