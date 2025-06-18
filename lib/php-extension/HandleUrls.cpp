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

    eventCache.outgoingRequestUrl = CallPhpFunctionCurlGetInfo(curlHandle, CURLINFO_EFFECTIVE_URL);
    if (eventCache.outgoingRequestUrl.empty()) return;

    eventId = EVENT_PRE_OUTGOING_REQUEST;
    eventCache.moduleName = "curl";
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
    eventCache.moduleName = "curl";
    eventCache.outgoingRequestEffectiveUrl = CallPhpFunctionCurlGetInfo(curlHandle, CURLINFO_EFFECTIVE_URL);
    eventCache.outgoingRequestPort = CallPhpFunctionCurlGetInfo(curlHandle, CURLINFO_PRIMARY_PORT);
    eventCache.outgoingRequestResolvedIp = CallPhpFunctionCurlGetInfo(curlHandle, CURLINFO_PRIMARY_IP);
}
