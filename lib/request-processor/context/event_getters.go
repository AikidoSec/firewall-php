package context

// #include "../../API.h"
import "C"
import (
	"main/helpers"
	"main/instance"
	"net/url"
)

func GetOutgoingRequestHostnameAndPort(inst *instance.RequestProcessorInstance) (string, uint32) {
	return getHostNameAndPort(inst, C.OUTGOING_REQUEST_URL, C.OUTGOING_REQUEST_PORT)
}

func GetOutgoingRequestEffectiveHostnameAndPort(inst *instance.RequestProcessorInstance) (string, uint32) {
	return getHostNameAndPort(inst, C.OUTGOING_REQUEST_EFFECTIVE_URL, C.OUTGOING_REQUEST_EFFECTIVE_URL_PORT)
}

func GetOutgoingRequestResolvedIp(inst *instance.RequestProcessorInstance) string {
	return GetContext(inst).Callback(inst, C.OUTGOING_REQUEST_RESOLVED_IP)
}

func GetFunctionName(inst *instance.RequestProcessorInstance) string {
	return GetContext(inst).Callback(inst, C.FUNCTION_NAME)
}

func GetCmd(inst *instance.RequestProcessorInstance) string {
	return GetContext(inst).Callback(inst, C.CMD)
}

func GetFilename(inst *instance.RequestProcessorInstance) string {
	return GetContext(inst).Callback(inst, C.FILENAME)
}

func GetFilename2(inst *instance.RequestProcessorInstance) string {
	return GetContext(inst).Callback(inst, C.FILENAME2)
}

func GetSqlQuery(inst *instance.RequestProcessorInstance) string {
	return GetContext(inst).Callback(inst, C.SQL_QUERY)
}

func GetSqlDialect(inst *instance.RequestProcessorInstance) string {
	return GetContext(inst).Callback(inst, C.SQL_DIALECT)
}

func GetModule(inst *instance.RequestProcessorInstance) string {
	return GetContext(inst).Callback(inst, C.MODULE)
}

func GetStackTrace(inst *instance.RequestProcessorInstance) string {
	return GetContext(inst).Callback(inst, C.STACK_TRACE)
}

func GetParamMatcher(inst *instance.RequestProcessorInstance) (string, string) {
	ctx := GetContext(inst)
	param := ctx.Callback(inst, C.PARAM_MATCHER_PARAM)
	regex := ctx.Callback(inst, C.PARAM_MATCHER_REGEX)
	return param, regex
}

func getHostNameAndPort(inst *instance.RequestProcessorInstance, urlCallbackId int, portCallbackId int) (string, uint32) {
	ctx := GetContext(inst)
	urlStr := ctx.Callback(inst, urlCallbackId)
	urlParsed, err := url.Parse(urlStr)
	if err != nil {
		return "", 0
	}
	hostname := urlParsed.Hostname()
	portFromURL := helpers.GetPortFromURL(urlParsed)

	portStr := ctx.Callback(inst, portCallbackId)
	port := helpers.ParsePort(portStr)
	if port == 0 {
		port = portFromURL
	}
	return hostname, port
}
