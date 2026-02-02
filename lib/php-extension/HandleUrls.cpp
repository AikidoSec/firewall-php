#include "Includes.h"

AIKIDO_HANDLER_FUNCTION(handle_pre_curl_exec) {
    scopedTimer.SetSink(sink, "outgoing_http_op");

    zval *curlHandle = NULL;

    ZEND_PARSE_PARAMETERS_START(1, 1)
#if PHP_VERSION_ID >= 80000
    Z_PARAM_OBJECT(curlHandle)
#else
    Z_PARAM_RESOURCE(curlHandle)
#endif
    ZEND_PARSE_PARAMETERS_END();

    eventCacheStack.Top().outgoingRequestUrl = CallPhpFunctionCurlGetInfo(curlHandle, CURLINFO_EFFECTIVE_URL);
    eventCacheStack.Top().outgoingRequestPort = CallPhpFunctionCurlGetInfo(curlHandle, CURLINFO_PRIMARY_PORT);

    // if requestCache.outgoingRequestUrl is not empty, we check if it's a redirect
    if (!requestCache.outgoingRequestUrl.empty()) {
        json outgoingRequestUrlJson = CallPhpFunctionParseUrl(eventCacheStack.Top().outgoingRequestUrl);
        json outgoingRequestRedirectUrlJson = CallPhpFunctionParseUrl(requestCache.outgoingRequestRedirectUrl);

        // if the host and port are the same, we use the initial URL, otherwise we use the effective URL
        if (!outgoingRequestUrlJson.empty() && !outgoingRequestRedirectUrlJson.empty() &&
            outgoingRequestUrlJson["host"] == outgoingRequestRedirectUrlJson["host"] && 
            outgoingRequestUrlJson["port"] == outgoingRequestRedirectUrlJson["port"]) {

            eventCacheStack.Top().outgoingRequestUrl = requestCache.outgoingRequestUrl;
        } else {
            // if previous outgoingRequestRedirectUrl it's different from outgoingRequestUrl it means that it's a new request 
            // so we reset the outgoingRequestUrl
            requestCache.outgoingRequestUrl = "";
        }
    }

    if (eventCacheStack.Top().outgoingRequestUrl.empty()) return;

    eventId = EVENT_PRE_OUTGOING_REQUEST;
    eventCacheStack.Top().moduleName = "curl";
}

AIKIDO_HANDLER_FUNCTION(handle_post_curl_exec) {
    zval *curlHandle = NULL;

// Curl handles changed between PHP 7 & PHP 8 - so we need different extraction
#if PHP_VERSION_ID >= 80000
    ZEND_PARSE_PARAMETERS_START(1, 1)
    Z_PARAM_OBJECT(curlHandle)
    ZEND_PARSE_PARAMETERS_END();
#else
    ZEND_PARSE_PARAMETERS_START(1, 1)
    Z_PARAM_RESOURCE(curlHandle)
    ZEND_PARSE_PARAMETERS_END();
#endif


    eventId = EVENT_POST_OUTGOING_REQUEST;
    eventCacheStack.Top().moduleName = "curl";
    eventCacheStack.Top().outgoingRequestEffectiveUrl = CallPhpFunctionCurlGetInfo(curlHandle, CURLINFO_EFFECTIVE_URL);
    eventCacheStack.Top().outgoingRequestEffectiveUrlPort = CallPhpFunctionCurlGetInfo(curlHandle, CURLINFO_PRIMARY_PORT);
    eventCacheStack.Top().outgoingRequestResolvedIp = CallPhpFunctionCurlGetInfo(curlHandle, CURLINFO_PRIMARY_IP);
    std::string outgoingRequestResponseCode = CallPhpFunctionCurlGetInfo(curlHandle, CURLINFO_RESPONSE_CODE);
    
    // if outgoingRequestResponseCode starts with 3, it's a redirect 
    if (!outgoingRequestResponseCode.empty() && outgoingRequestResponseCode.substr(0, 1) == "3") {
        requestCache.outgoingRequestRedirectUrl = CallPhpFunctionCurlGetInfo(curlHandle, CURLINFO_REDIRECT_URL);  
  
        // if it's the first redirect
        if (requestCache.outgoingRequestUrl.empty()) {
            requestCache.outgoingRequestUrl = eventCacheStack.Top().outgoingRequestEffectiveUrl;
        }
    } 
    else {
        requestCache.outgoingRequestUrl = "";
        requestCache.outgoingRequestRedirectUrl = "";
    }
    
}
