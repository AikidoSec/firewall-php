package context

// #include "../../API.h"
import "C"
import (
	"main/helpers"
	"net/url"
)

func GetOutgoingRequestHostnameAndPort() (string, uint32) {
	return getHostNameAndPort(C.OUTGOING_REQUEST_URL, C.OUTGOING_REQUEST_PORT)
}

func GetOutgoingRequestEffectiveHostnameAndPort() (string, uint32) {
	return getHostNameAndPort(C.OUTGOING_REQUEST_EFFECTIVE_URL, C.OUTGOING_REQUEST_EFFECTIVE_URL_PORT)
}

func GetOutgoingRequestResolvedIp() string {
	return Context.Callback(C.OUTGOING_REQUEST_RESOLVED_IP)
}

func GetFunctionName() string {
	return Context.Callback(C.FUNCTION_NAME)
}

func GetCmd() string {
	return Context.Callback(C.CMD)
}

func GetFilename() string {
	return Context.Callback(C.FILENAME)
}

func GetFilename2() string {
	return Context.Callback(C.FILENAME2)
}

func GetSqlQuery() string {
	return Context.Callback(C.SQL_QUERY)
}

func GetSqlDialect() string {
	return Context.Callback(C.SQL_DIALECT)
}

func GetSqlParams() string {
	return Context.Callback(C.SQL_PARAMS)
}

func GetModule() string {
	return Context.Callback(C.MODULE)
}

func GetStackTrace() string {
	return Context.Callback(C.STACK_TRACE)
}

func GetParamMatcher() (string, string) {
	param := Context.Callback(C.PARAM_MATCHER_PARAM)
	regex := Context.Callback(C.PARAM_MATCHER_REGEX)
	return param, regex
}

func GetTenantId() string {
	return Context.Callback(C.CONTEXT_TENANT_ID)
}

func IsIdorDisabled() bool {
	return Context.Callback(C.CONTEXT_IDOR_DISABLED) == "1"
}

func GetIdorTenantColumnName() string {
	return Context.Callback(C.CONTEXT_IDOR_TENANT_COLUMN_NAME)
}

func GetIdorExcludedTables() string {
	return Context.Callback(C.CONTEXT_IDOR_EXCLUDED_TABLES)
}

func getHostNameAndPort(urlCallbackId int, portCallbackId int) (string, uint32) { // urlcallbackid is the type of data we request, eg C.OUTGOING_REQUEST_URL
	urlStr := Context.Callback(urlCallbackId)
	urlParsed, err := url.Parse(urlStr)
	if err != nil {
		return "", 0
	}
	hostname := urlParsed.Hostname()
	portFromURL := helpers.GetPortFromURL(urlParsed)

	portStr := Context.Callback(portCallbackId)
	port := helpers.ParsePort(portStr)
	if port == 0 {
		port = portFromURL
	}
	return hostname, port
}
