#include "Includes.h"

GoString GoCreateString(const std::string& s) {
    return GoString{s.c_str(), s.length()};
}

GoSlice GoCreateSlice(const std::vector<int64_t>& v) {
    return GoSlice{ (void*)v.data(), v.size(), v.capacity() };
}
/*
    Callback wrapper called by the RequestProcessor (GO) whenever it needs data from PHP (C++ extension).
*/
char* GoContextCallback(int callbackId) {
    if (!request.Ok()) {
        AIKIDO_LOG_WARN("Request is not ok!\n");
        return nullptr;
    }
    std::string ctx;
    std::string ret;

    try {
        switch (callbackId) {
            case CONTEXT_REMOTE_ADDRESS:
                ctx = "REMOTE_ADDRESS";
                ret = request.GetVar("REMOTE_ADDR");
                break;
            case CONTEXT_METHOD:
                ctx = "METHOD";
                ret = request.GetVar("REQUEST_METHOD");
                break;
            case CONTEXT_ROUTE:
                ctx = "ROUTE";
                ret = request.GetRoute();
                break;
            case CONTEXT_STATUS_CODE:
                ctx = "STATUS_CODE";
                ret = request.GetStatusCode();
                break;
            case CONTEXT_BODY:
                ctx = "BODY";
                ret = request.GetBody();
                break;
            case CONTEXT_HEADER_X_FORWARDED_FOR:
                ctx = "HEADER_X_FORWARDED_FOR";
                ret = request.GetVar("HTTP_X_FORWARDED_FOR");
                break;
            case CONTEXT_COOKIES:
                ctx = "COOKIES";
                ret = request.GetVar("HTTP_COOKIE");
                break;
            case CONTEXT_QUERY:
                ctx = "QUERY";
                ret = request.GetQuery();
                break;
            case CONTEXT_HTTPS:
                ctx = "HTTPS";
                ret = request.GetVar("HTTPS");
                break;
            case CONTEXT_URL:
                ctx = "URL";
                ret = request.GetUrl();
                break;
            case CONTEXT_HEADERS:
                ctx = "HEADERS";
                ret = request.GetHeaders();
                break;
            case CONTEXT_HEADER_USER_AGENT:
                ctx = "USER_AGENT";
                ret = request.GetVar("HTTP_USER_AGENT");
                break;
            case CONTEXT_USER_ID:
                ctx = "USER_ID";
                ret = requestCache.userId;
                break;
            case CONTEXT_USER_NAME:
                ctx = "USER_NAME";
                ret = requestCache.userName;
                break;
            case FUNCTION_NAME:
                ctx = "FUNCTION_NAME";
                ret = eventCache.functionName;
                break;
            case OUTGOING_REQUEST_URL:
                ctx = "OUTGOING_REQUEST_URL";
                ret = eventCache.outgoingRequestUrl;
                break;
            case OUTGOING_REQUEST_EFFECTIVE_URL:
                ctx = "OUTGOING_REQUEST_EFFECTIVE_URL";
                ret = eventCache.outgoingRequestEffectiveUrl;
                break;
            case OUTGOING_REQUEST_PORT:
                ctx = "OUTGOING_REQUEST_PORT";
                ret = eventCache.outgoingRequestPort;
                break;
            case OUTGOING_REQUEST_RESOLVED_IP:
                ctx = "OUTGOING_REQUEST_RESOLVED_IP";
                ret = eventCache.outgoingRequestResolvedIp;
                break;
            case CMD:
                ctx = "CMD";
                ret = eventCache.cmd;
                break;
            case FILENAME:
                ctx = "FILENAME";
                ret = eventCache.filename;
                break;
            case FILENAME2:
                ctx = "FILENAME2";
                ret = eventCache.filename2;
                break;
            case SQL_QUERY:
                ctx = "SQL_QUERY";
                ret = eventCache.sqlQuery;
                break;
            case SQL_DIALECT:
                ctx = "SQL_DIALECT";
                ret = eventCache.sqlDialect;
                break;
            case MODULE:
                ctx = "MODULE";
                ret = eventCache.moduleName;
                break;
        }
    } catch (std::exception& e) {
        AIKIDO_LOG_ERROR("Exception in GoContextCallback: %s\n", e.what());
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
