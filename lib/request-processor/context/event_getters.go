package context

// #include "../../API.h"
import "C"
import (
	"main/helpers"
	"net/url"
)

func GetOutgoingRequestHostnameAndPort() (string, uint32) {
	return getHostNameAndPort(C.OUTGOING_REQUEST_URL)
}

func GetOutgoingRequestEffectiveHostnameAndPort() (string, uint32) {
	return getHostNameAndPort(C.OUTGOING_REQUEST_EFFECTIVE_URL)
}

func GetOutgoingRequestResolvedIp() string {
	return Context.Callback(C.OUTGOING_REQUEST_RESOLVED_IP)
}

func GetOutgoingRequestResponseCode() string {
	return Context.Callback(C.OUTGOING_REQUEST_RESPONSE_CODE)
}

func GetOutgoingRequestRedirectUrl() string {
	host, _ := getHostNameAndPort(C.OUTGOING_REQUEST_REDIRECT_URL)
	return host
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

func GetModule() string {
	return Context.Callback(C.MODULE)
}

func getHostNameAndPort(urlCallbackId int) (string, uint32) { // urlcallbackid is the type of data we request, eg C.OUTGOING_REQUEST_URL
	urlStr := Context.Callback(urlCallbackId)
	// remove all control characters (< 32) and 0x7f(DEL) also replace \@ with @ and remove all whitespace
	// url.Parse fails if the url contains control characters
	urlStr = helpers.NormalizeRawUrl(urlStr)
	urlParsed, err := url.Parse(urlStr)
	if err != nil {
		return "", 0
	}
	hostname := urlParsed.Hostname()
	portFromURL := helpers.GetPortFromURL(urlParsed)

	portStr := Context.Callback(C.OUTGOING_REQUEST_PORT)
	port := helpers.ParsePort(portStr)
	if port == 0 {
		port = portFromURL
	}
	return hostname, port
}
