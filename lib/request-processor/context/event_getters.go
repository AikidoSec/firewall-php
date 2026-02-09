package context

// #include "../../API.h"
import "C"
import (
	"main/helpers"
	"main/instance"
	"net/url"
)

func GetOutgoingRequestHostnameAndPort(instance *instance.RequestProcessorInstance) (string, uint32) {
	return getHostNameAndPort(instance, C.OUTGOING_REQUEST_URL, C.OUTGOING_REQUEST_PORT)
}

func GetOutgoingRequestEffectiveHostnameAndPort(instance *instance.RequestProcessorInstance) (string, uint32) {
	return getHostNameAndPort(instance, C.OUTGOING_REQUEST_EFFECTIVE_URL, C.OUTGOING_REQUEST_EFFECTIVE_URL_PORT)
}

func GetOutgoingRequestResolvedIp(instance *instance.RequestProcessorInstance) string {
	return GetContext(instance).Callback(instance, C.OUTGOING_REQUEST_RESOLVED_IP)
}

func GetFunctionName(instance *instance.RequestProcessorInstance) string {
	return GetContext(instance).Callback(instance, C.FUNCTION_NAME)
}

func GetCmd(instance *instance.RequestProcessorInstance) string {
	return GetContext(instance).Callback(instance, C.CMD)
}

func GetFilename(instance *instance.RequestProcessorInstance) string {
	return GetContext(instance).Callback(instance, C.FILENAME)
}

func GetFilename2(instance *instance.RequestProcessorInstance) string {
	return GetContext(instance).Callback(instance, C.FILENAME2)
}

func GetSqlQuery(instance *instance.RequestProcessorInstance) string {
	return GetContext(instance).Callback(instance, C.SQL_QUERY)
}

func GetSqlDialect(instance *instance.RequestProcessorInstance) string {
	return GetContext(instance).Callback(instance, C.SQL_DIALECT)
}

func GetModule(instance *instance.RequestProcessorInstance) string {
	return GetContext(instance).Callback(instance, C.MODULE)
}

func GetStackTrace(instance *instance.RequestProcessorInstance) string {
	return GetContext(instance).Callback(instance, C.STACK_TRACE)
}

func GetParamMatcher(instance *instance.RequestProcessorInstance) (string, string) {
	ctx := GetContext(instance)
	param := ctx.Callback(instance, C.PARAM_MATCHER_PARAM)
	regex := ctx.Callback(instance, C.PARAM_MATCHER_REGEX)
	return param, regex
}

func getHostNameAndPort(instance *instance.RequestProcessorInstance, urlCallbackId int, portCallbackId int) (string, uint32) {
	ctx := GetContext(instance)
	urlStr := ctx.Callback(instance, urlCallbackId)
	urlParsed, err := url.Parse(urlStr)
	if err != nil {
		return "", 0
	}
	hostname := urlParsed.Hostname()
	portFromURL := helpers.GetPortFromURL(urlParsed)

	portStr := ctx.Callback(instance, portCallbackId)
	port := helpers.ParsePort(portStr)
	if port == 0 {
		port = portFromURL
	}
	return hostname, port
}
