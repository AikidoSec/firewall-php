#include "Includes.h"

GoString GoCreateString(const std::string& s) {
    return GoString{ s.c_str(), static_cast<ptrdiff_t>(s.size()) };
}

GoSlice GoCreateSlice(const std::vector<int64_t>& v) {
    return GoSlice{ static_cast<void*>(const_cast<int64_t*>(v.data())),
                    static_cast<GoInt>(v.size()),
                    static_cast<GoInt>(v.capacity()) };
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
                ret = AIKIDO_GLOBAL(server).GetVar("REMOTE_ADDR");
                break;
            case CONTEXT_METHOD:
                ctx = "METHOD";
                ret = AIKIDO_GLOBAL(server).GetVar("REQUEST_METHOD");
                break;
            case CONTEXT_ROUTE:
                ctx = "ROUTE";
                ret = AIKIDO_GLOBAL(server).GetRoute();
                break;
            case CONTEXT_STATUS_CODE:
                ctx = "STATUS_CODE";
                ret = AIKIDO_GLOBAL(server).GetStatusCode();
                break;
            case CONTEXT_BODY:
                ctx = "BODY";
                ret = AIKIDO_GLOBAL(server).GetBody();
                break;
            case CONTEXT_HEADER_X_FORWARDED_FOR:
                ctx = "HEADER_X_FORWARDED_FOR";
                ret = AIKIDO_GLOBAL(server).GetVar("HTTP_X_FORWARDED_FOR");
                break;
            case CONTEXT_COOKIES:
                ctx = "COOKIES";
                ret = AIKIDO_GLOBAL(server).GetVar("HTTP_COOKIE");
                break;
            case CONTEXT_QUERY:
                ctx = "QUERY";
                ret = AIKIDO_GLOBAL(server).GetQuery();
                break;
            case CONTEXT_HTTPS:
                ctx = "HTTPS";
                ret = AIKIDO_GLOBAL(server).GetVar("HTTPS");
                break;
            case CONTEXT_URL:
                ctx = "URL";
                ret = AIKIDO_GLOBAL(server).GetUrl();
                break;
            case CONTEXT_HEADERS:
                ctx = "HEADERS";
                ret = AIKIDO_GLOBAL(server).GetHeaders();
                break;
            case CONTEXT_HEADER_USER_AGENT:
                ctx = "USER_AGENT";
                ret = AIKIDO_GLOBAL(server).GetVar("HTTP_USER_AGENT");
                break;
            case CONTEXT_USER_ID:
                ctx = "USER_ID";
                ret = AIKIDO_GLOBAL(requestCache).userId;
                break;
            case CONTEXT_USER_NAME:
                ctx = "USER_NAME";
                ret = AIKIDO_GLOBAL(requestCache).userName;
                break;
            case CONTEXT_RATE_LIMIT_GROUP:
                ctx = "RATE_LIMIT_GROUP";
                ret = AIKIDO_GLOBAL(requestCache).rateLimitGroup;
                break;
            case FUNCTION_NAME:
                ctx = "FUNCTION_NAME";
                ret = AIKIDO_GLOBAL(eventCache).functionName;
                break;
            case OUTGOING_REQUEST_URL:
                ctx = "OUTGOING_REQUEST_URL";
                ret = AIKIDO_GLOBAL(eventCache).outgoingRequestUrl;
                break;
            case OUTGOING_REQUEST_EFFECTIVE_URL:
                ctx = "OUTGOING_REQUEST_EFFECTIVE_URL";
                ret = AIKIDO_GLOBAL(eventCache).outgoingRequestEffectiveUrl;
                break;
            case OUTGOING_REQUEST_PORT:
                ctx = "OUTGOING_REQUEST_PORT";
                ret = AIKIDO_GLOBAL(eventCache).outgoingRequestPort;
                break;
            case OUTGOING_REQUEST_EFFECTIVE_URL_PORT:
                ctx = "OUTGOING_REQUEST_EFFECTIVE_URL_PORT";
                ret = eventCache.outgoingRequestEffectiveUrlPort;
                break;
            case OUTGOING_REQUEST_RESOLVED_IP:
                ctx = "OUTGOING_REQUEST_RESOLVED_IP";
                ret = AIKIDO_GLOBAL(eventCache).outgoingRequestResolvedIp;
                break;
            case CMD:
                ctx = "CMD";
                ret = AIKIDO_GLOBAL(eventCache).cmd;
                break;
            case FILENAME:
                ctx = "FILENAME";
                ret = AIKIDO_GLOBAL(eventCache).filename;
                break;
            case FILENAME2:
                ctx = "FILENAME2";
                ret = AIKIDO_GLOBAL(eventCache).filename2;
                break;
            case SQL_QUERY:
                ctx = "SQL_QUERY";
                ret = AIKIDO_GLOBAL(eventCache).sqlQuery;
                break;
            case SQL_DIALECT:
                ctx = "SQL_DIALECT";
                ret = AIKIDO_GLOBAL(eventCache).sqlDialect;
                break;
            case MODULE:
                ctx = "MODULE";
                ret = AIKIDO_GLOBAL(eventCache).moduleName;
                break;
            case STACK_TRACE:
                ctx = "STACK_TRACE";
                ret = GetStackTrace();
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
