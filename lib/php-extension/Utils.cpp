#include "Includes.h"

std::string ToLowercase(const std::string& str) {
    std::string result = str;
    std::transform(result.begin(), result.end(), result.begin(), [](unsigned char c) { return std::tolower(c); });
    return result;
}

std::string ToUppercase(const std::string& str) {
    std::string result = str;
    std::transform(result.begin(), result.end(), result.begin(), [](unsigned char c) { return std::toupper(c); });
    return result;
}

std::string GetRandomNumber() {
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> dis(100000, 999999);
    return std::to_string(int(dis(gen)));
}

std::string GetTime() {
    std::time_t current_time = std::time(nullptr);
    char time_str[20];
    std::strftime(time_str, sizeof(time_str), "%H:%M:%S", std::localtime(&current_time));
    return time_str;
}

std::string GetDateTime() {
    std::time_t current_time = std::time(nullptr);
    char time_str[20];
    std::strftime(time_str, sizeof(time_str), "%Y%m%d%H%M%S", std::localtime(&current_time));
    return time_str;
}

pid_t GetThreadID() {
#ifdef SYS_gettid
    return syscall(SYS_gettid);
#else
    return (pid_t)getpid(); // Fallback for non-Linux systems
#endif
}
const char* GetEventName(EVENT_ID event) {
    switch (event) {
        case EVENT_PRE_REQUEST:
            return "PreRequest";
        case EVENT_POST_REQUEST:
            return "PostRequest";
        case EVENT_SET_USER:
            return "SetUser";
        case EVENT_GET_AUTO_BLOCKING_STATUS:
            return "GetAutoBlockingStatus";
        case EVENT_GET_BLOCKING_STATUS:
            return "GetBlockingStatus";
        case EVENT_PRE_OUTGOING_REQUEST:
            return "PreOutgoingRequest";
        case EVENT_POST_OUTGOING_REQUEST:
            return "PostOutgoingRequest";
        case EVENT_PRE_SHELL_EXECUTED:
            return "PreShellExecuted";
        case EVENT_PRE_PATH_ACCESSED:
            return "PrePathAccessed";
        case EVENT_PRE_SQL_QUERY_EXECUTED:
            return "PreSqlQueryExecuted";
    }
    return "Unknown";
}

std::string NormalizeAndDumpJson(const json& jsonObj) {
    // Remove invalid UTF8 characters (normalize)
    // https://json.nlohmann.me/api/basic_json/dump/
    return jsonObj.dump(-1, ' ', false, json::error_handler_t::ignore);
}

std::string ArrayToJson(zval* array) {
    if (!array) {
        return "";
    }

    json query_json;
    zend_string *key;
    zval *val;
    ZEND_HASH_FOREACH_STR_KEY_VAL(Z_ARRVAL_P(array), key, val) {
        if(key && val) {
            std::string key_str(ZSTR_VAL(key));
            if (Z_TYPE_P(val) == IS_STRING) {
                query_json[key_str] = Z_STRVAL_P(val);
            }
            else if (Z_TYPE_P(val) == IS_ARRAY){
                json val_array = json::array();
                zval *v;
                ZEND_HASH_FOREACH_VAL(Z_ARRVAL_P(val), v) {
                    if (Z_TYPE_P(v) == IS_STRING) {
                        val_array.push_back(Z_STRVAL_P(v));
                    }
                }
                ZEND_HASH_FOREACH_END();
                query_json[key_str] = val_array;
            }
        }
    }
    ZEND_HASH_FOREACH_END();

    return NormalizeAndDumpJson(query_json);
}

std::string GetSqlDialectFromPdo(zval *pdo_object) {
    if (!pdo_object) {
        return "unknown";
    }

    zval retval;
    std::string result = "unknown";
    if (CallPhpFunctionWithOneParam("getAttribute", PDO_ATTR_DRIVER_NAME, &retval, pdo_object)) {
        if (Z_TYPE(retval) == IS_STRING) {
            result = Z_STRVAL(retval);
        }
    }
    zval_ptr_dtor(&retval);
    return result;
}

bool StartsWith(const std::string& str, const std::string& prefix, bool caseSensitive) {
    std::string strToCompare = str;
    std::string prefixToCompare = prefix;
    if (!caseSensitive) {
        strToCompare = ToLowercase(str);
        prefixToCompare = ToLowercase(prefix);
    }
    return strToCompare.size() >= prefixToCompare.size() && strToCompare.compare(0, prefixToCompare.length(), prefixToCompare) == 0;
}

json CallPhpFunctionParseUrl(const std::string& url) {
    if (url.empty()) {
        return json();
    }

    zval retval;
    json result_json;
    if (CallPhpFunctionWithOneParam("parse_url", url, &retval)) {
        if (Z_TYPE(retval) == IS_ARRAY) {
            zval* host = zend_hash_str_find(Z_ARRVAL(retval), "host", sizeof("host") - 1);
            if (host && Z_TYPE_P(host) == IS_STRING) {
                result_json["host"] = Z_STRVAL_P(host);
            }
           
            zval* port = zend_hash_str_find(Z_ARRVAL(retval), "port", sizeof("port") - 1);
            if (port && Z_TYPE_P(port) == IS_LONG) {
                result_json["port"] = Z_LVAL_P(port);
            } else {
                zval* scheme = zend_hash_str_find(Z_ARRVAL(retval), "scheme", sizeof("scheme") - 1);
                if (scheme && Z_TYPE_P(scheme) == IS_STRING) {
                    if (strcmp(Z_STRVAL_P(scheme), "https") == 0) {
                        result_json["port"] = 443;
                    } 
                    else if (strcmp(Z_STRVAL_P(scheme), "http") == 0) {
                        result_json["port"] = 80;
                    } 
                    else {
                        result_json["port"] = 0;
                    }
                }
            }
        }
    }
    zval_ptr_dtor(&retval);
    return result_json;
}

std::string AnonymizeToken(const std::string& str) {
    return str.length() > 4 ? "AIK_RUNTIME_***" + str.substr(str.length() - 4) : "AIK_RUNTIME_***";
}

bool FileExists(const std::string& filePath) {
    struct stat buffer;
    if (stat(filePath.c_str(), &buffer) == 0) {
        return true;
    }
    return false;
}

bool RemoveFile(const std::string& filePath) {
    if (unlink(filePath.c_str()) == 0) {
        return true;
    }
    return false;
}


std::string GetStackTrace() {
#if PHP_VERSION_ID >= 80100
    // Check if there's an active execution context
    if (!EG(current_execute_data)) {
        return "";
    }

    zval trace;
    zend_fetch_debug_backtrace(&trace, 0, DEBUG_BACKTRACE_IGNORE_ARGS, 0);

    if (Z_TYPE(trace) != IS_ARRAY) {
        zval_ptr_dtor(&trace);
        return "";
    }

    zend_string *trace_string = zend_trace_to_string(Z_ARRVAL(trace), true);

    std::string result;
    if (trace_string) {
        result = std::string(ZSTR_VAL(trace_string), ZSTR_LEN(trace_string));
        zend_string_release(trace_string);
    }

    zval_ptr_dtor(&trace);
    return result;
#else
    return "";
#endif
}
