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

AIKIDO_HANDLER_FUNCTION(handle_pre_socket_connect) {
    scopedTimer.SetSink(sink, "outgoing_http_op");

    zval *socketHandle = NULL;
    zval *address = NULL;
    zval *port = NULL;
    php_socket *phpSock = NULL;

    ZEND_PARSE_PARAMETERS_START(0, -1)
    Z_PARAM_OPTIONAL
#if PHP_VERSION_ID >= 80000
    Z_PARAM_OBJECT(socketHandle)
#else
    Z_PARAM_RESOURCE(socketHandle)
#endif
    Z_PARAM_ZVAL(address)
#if PHP_VERSION_ID >= 80000
    Z_PARAM_ZVAL_OR_NULL(port)
#else
    Z_PARAM_ZVAL_EX(port, 0, 1)
#endif
    ZEND_PARSE_PARAMETERS_END();
    
#if PHP_VERSION_ID >= 80000
    if (socketHandle) {
        phpSock = Z_SOCKET_P(socketHandle);
        ENSURE_SOCKET_VALID(phpSock);
            // if the socket is not an IP address, we return
            if (phpSock->type != AF_INET && phpSock->type != AF_INET6) {
                return;
            }
    }
#else
    // For PHP 7, we can't access the socket resource type directly
    // as php_sockets_le_socket() is not exported. We'll rely on address validation instead.
#endif

    std::string addressStr = "";
    std::string portStr = "";

    if (address && Z_TYPE_P(address) == IS_STRING) {
        addressStr = Z_STRVAL_P(address);
    } else if (address && Z_TYPE_P(address) == IS_LONG) {
        // If address is numeric, it might be an IP address
        addressStr = std::to_string(Z_LVAL_P(address));
    }

    if (port && Z_TYPE_P(port) == IS_LONG && Z_LVAL_P(port) > 0) {
        portStr = std::to_string(Z_LVAL_P(port));
    } else if (port && Z_TYPE_P(port) == IS_STRING) {
        portStr = Z_STRVAL_P(port);
    }

    if (addressStr.empty()) {
        return;
    }

    if (!portStr.empty()) {
            eventCache.outgoingRequestUrl = "tcp://" + addressStr + ":" + portStr;
            eventCache.outgoingRequestPort = portStr;
        } else {
            eventCache.outgoingRequestUrl = "tcp://" + addressStr;
            eventCache.outgoingRequestPort = "80"; 
    }

    eventId = EVENT_PRE_OUTGOING_REQUEST;
}

AIKIDO_HANDLER_FUNCTION(handle_post_socket_connect) {
    eventId = EVENT_POST_OUTGOING_REQUEST;
    // For socket_connect, we don't have easy access to resolved IP after connection
    // The URL was already set in pre handler
    eventCache.outgoingRequestEffectiveUrl = eventCache.outgoingRequestUrl;
    eventCache.outgoingRequestEffectiveUrlPort = eventCache.outgoingRequestPort;
}

AIKIDO_HANDLER_FUNCTION(handle_pre_fsockopen) {
    scopedTimer.SetSink(sink, "outgoing_http_op");

    zval *hostname = NULL;
    zend_long port = -1;

    ZEND_PARSE_PARAMETERS_START(0, -1)
    Z_PARAM_OPTIONAL
    Z_PARAM_ZVAL(hostname)
    Z_PARAM_LONG(port)
    ZEND_PARSE_PARAMETERS_END();

    std::string hostnameStr = "";
    std::string portStr = "";

    if (hostname && Z_TYPE_P(hostname) == IS_STRING) {
        hostnameStr = Z_STRVAL_P(hostname);
    }

    if (port >= 0) {
        portStr = std::to_string(port);
    }

    if (!hostnameStr.empty()) {
        if (!portStr.empty()) {
            eventCache.outgoingRequestUrl = "tcp://" + hostnameStr + ":" + portStr;
            eventCache.outgoingRequestPort = portStr;
        } else {
            eventCache.outgoingRequestUrl = "tcp://" + hostnameStr;
            eventCache.outgoingRequestPort = "80"; // Default port
        }
    }

    if (eventCache.outgoingRequestUrl.empty()) return;

    eventId = EVENT_PRE_OUTGOING_REQUEST;
}

AIKIDO_HANDLER_FUNCTION(handle_post_fsockopen) {
    eventId = EVENT_POST_OUTGOING_REQUEST;
    // For fsockopen, we don't have easy access to resolved IP after connection
    // The URL was already set in pre handler
    eventCache.outgoingRequestEffectiveUrl = eventCache.outgoingRequestUrl;
    eventCache.outgoingRequestEffectiveUrlPort = eventCache.outgoingRequestPort;
}

AIKIDO_HANDLER_FUNCTION(handle_pre_stream_socket_client) {
    scopedTimer.SetSink(sink, "outgoing_http_op");

    zval *address = NULL;

    ZEND_PARSE_PARAMETERS_START(0, -1)
    Z_PARAM_OPTIONAL
    Z_PARAM_ZVAL(address)
    ZEND_PARSE_PARAMETERS_END();

    std::string addressStr = "";

    if (address && Z_TYPE_P(address) == IS_STRING) {
        addressStr = Z_STRVAL_P(address);
    }

    if (addressStr.empty()){
        return;
    }

    eventCache.outgoingRequestUrl = addressStr;
    eventId = EVENT_PRE_OUTGOING_REQUEST;
}

AIKIDO_HANDLER_FUNCTION(handle_post_stream_socket_client) {
    eventId = EVENT_POST_OUTGOING_REQUEST;
    // For stream_socket_client, we don't have easy access to resolved IP after connection
    // The URL was already set in pre handler
    eventCache.outgoingRequestEffectiveUrl = eventCache.outgoingRequestUrl;
    eventCache.outgoingRequestEffectiveUrlPort = eventCache.outgoingRequestPort;
}
