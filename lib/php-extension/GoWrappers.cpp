#include "Includes.h"

GoString GoCreateString(const std::string& s) {
    return GoString{s.c_str(), s.length()};
}

GoSlice GoCreateSlice(const std::vector<int64_t>& v) {
    return GoSlice{ (void*)v.data(), v.size(), v.capacity() };
}

/*
    Helper function to safely get a string field from EventCache.
    Returns empty string if stack is empty, otherwise returns the field value.
*/
static inline std::string GetEventCacheField(std::string EventCache::*field) {
    return eventCacheStack.Empty() ? "" : eventCacheStack.Top().*field;
}

/*
    Callback wrapper called by the RequestProcessor (GO) whenever it needs data from PHP (C++ extension).
*/
char* GoContextCallback(int callbackId) {
    std::string ctx;
    std::string ret;

    try {
        switch (callbackId) {
            case CONTEXT_REMOTE_ADDRESS:
                ctx = "REMOTE_ADDRESS";
                ret = server.GetVar("REMOTE_ADDR");
                break;
            case CONTEXT_METHOD:
                ctx = "METHOD";
                ret = server.GetMethod();
                break;
            case CONTEXT_ROUTE:
                ctx = "ROUTE";
                ret = server.GetRoute();
                break;
            case CONTEXT_STATUS_CODE:
                ctx = "STATUS_CODE";
                ret = server.GetStatusCode();
                break;
            case CONTEXT_BODY:
                ctx = "BODY";
                ret = server.GetBody();
                break;
            case CONTEXT_HEADER_X_FORWARDED_FOR:
                ctx = "HEADER_X_FORWARDED_FOR";
                ret = server.GetVar("HTTP_X_FORWARDED_FOR");
                break;
            case CONTEXT_COOKIES:
                ctx = "COOKIES";
                ret = server.GetVar("HTTP_COOKIE");
                break;
            case CONTEXT_QUERY:
                ctx = "QUERY";
                ret = server.GetQuery();
                break;
            case CONTEXT_HTTPS:
                ctx = "HTTPS";
                ret = server.GetVar("HTTPS");
                break;
            case CONTEXT_URL:
                ctx = "URL";
                ret = server.GetUrl();
                break;
            case CONTEXT_HEADERS:
                ctx = "HEADERS";
                ret = server.GetHeaders();
                break;
            case CONTEXT_HEADER_USER_AGENT:
                ctx = "USER_AGENT";
                ret = server.GetVar("HTTP_USER_AGENT");
                break;
            case CONTEXT_USER_ID:
                ctx = "USER_ID";
                ret = requestCache.userId;
                break;
            case CONTEXT_USER_NAME:
                ctx = "USER_NAME";
                ret = requestCache.userName;
                break;
            case CONTEXT_RATE_LIMIT_GROUP:
                ctx = "RATE_LIMIT_GROUP";
                ret = requestCache.rateLimitGroup;
                break;
            case FUNCTION_NAME:
                ctx = "FUNCTION_NAME";
                ret = GetEventCacheField(&EventCache::functionName);
                break;
            case OUTGOING_REQUEST_URL:
                ctx = "OUTGOING_REQUEST_URL";
                ret = GetEventCacheField(&EventCache::outgoingRequestUrl);
                break;
            case OUTGOING_REQUEST_EFFECTIVE_URL:
                ctx = "OUTGOING_REQUEST_EFFECTIVE_URL";
                ret = GetEventCacheField(&EventCache::outgoingRequestEffectiveUrl);
                break;
            case OUTGOING_REQUEST_PORT:
                ctx = "OUTGOING_REQUEST_PORT";
                ret = GetEventCacheField(&EventCache::outgoingRequestPort);
                break;
            case OUTGOING_REQUEST_EFFECTIVE_URL_PORT:
                ctx = "OUTGOING_REQUEST_EFFECTIVE_URL_PORT";
                ret = GetEventCacheField(&EventCache::outgoingRequestEffectiveUrlPort);
                break;
            case OUTGOING_REQUEST_RESOLVED_IP:
                ctx = "OUTGOING_REQUEST_RESOLVED_IP";
                ret = GetEventCacheField(&EventCache::outgoingRequestResolvedIp);
                break;
            case CMD:
                ctx = "CMD";
                ret = GetEventCacheField(&EventCache::cmd);
                break;
            case FILENAME:
                ctx = "FILENAME";
                ret = GetEventCacheField(&EventCache::filename);
                break;
            case FILENAME2:
                ctx = "FILENAME2";
                ret = GetEventCacheField(&EventCache::filename2);
                break;
            case SQL_QUERY:
                ctx = "SQL_QUERY";
                ret = GetEventCacheField(&EventCache::sqlQuery);
                break;
            case SQL_DIALECT:
                ctx = "SQL_DIALECT";
                ret = GetEventCacheField(&EventCache::sqlDialect);
                break;
            case SQL_PARAMS:
                ctx = "SQL_PARAMS";
                ret = GetEventCacheField(&EventCache::sqlParams);
                break;
            case MODULE:
                ctx = "MODULE";
                ret = GetEventCacheField(&EventCache::moduleName);
                break;
            case STACK_TRACE:
                ctx = "STACK_TRACE";
                ret = GetStackTrace();
                break;
            case PARAM_MATCHER_PARAM:
                ctx = "PARAM_MATCHER_PARAM";
                ret = GetEventCacheField(&EventCache::paramMatcherParam);
                break;
            case PARAM_MATCHER_REGEX:
                ctx = "PARAM_MATCHER_REGEX";
                ret = GetEventCacheField(&EventCache::paramMatcherRegex);
                break;
            case CONTEXT_TENANT_ID:
                ctx = "TENANT_ID";
                ret = requestCache.tenantId;
                break;
            case CONTEXT_IDOR_DISABLED:
                ctx = "IDOR_DISABLED";
                ret = requestCache.idorDisabled ? "1" : "";
                break;
            case CONTEXT_IDOR_CONFIG:
                ctx = "IDOR_CONFIG";
                ret = requestCache.idorConfigJson;
                break;
        }
    } catch (std::exception& e) {
        AIKIDO_LOG_DEBUG("Exception in GoContextCallback: %s\n", e.what());
    }

    if (!ret.length()) {
        AIKIDO_LOG_DEBUG("Callback %s -> NULL\n", ctx.c_str());
        return nullptr;
    }

    if (ret.length() > 10000) {
        AIKIDO_LOG_DEBUG("Callback %s -> (Result too large to print)\n", ctx.c_str());
    } else {
        AIKIDO_LOG_DEBUG("Callback %s -> %s\n", ctx.c_str(), ret.c_str());
    }
    return strdup(ret.c_str());
}
