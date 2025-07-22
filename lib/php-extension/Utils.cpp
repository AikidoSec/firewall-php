#include "Includes.h"

std::string ToLowercase(const std::string& str) {
    std::string result = str;
    std::transform(result.begin(), result.end(), result.begin(), [](unsigned char c) { return std::tolower(c); });
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

std::string GenerateSocketPath() {
    return "/run/aikido-" + std::string(PHP_AIKIDO_VERSION) + "/aikido-" + GetDateTime() + "-" + GetRandomNumber() + ".sock";
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
    if (CallPhpFunctionWithOneParam("getAttribute", PDO_ATTR_DRIVER_NAME, &retval, pdo_object)) {
        if (Z_TYPE(retval) == IS_STRING) {
            return Z_STRVAL(retval);
        }
    }
    return "unknown";
}

bool StartsWith(const std::string& str, const std::string& prefix) {
    return str.size() >= prefix.size() && str.compare(0, prefix.length(), prefix) == 0;
}