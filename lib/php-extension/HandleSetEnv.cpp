#include "Includes.h"

ZEND_FUNCTION(set_env) {
    ScopedTimer scopedTimer("set_env", "aikido_op");

    if (AIKIDO_GLOBAL(disable) == true) {
        RETURN_BOOL(false);
    }

    char* key = nullptr;
    size_t keyLength = 0;
    char* value = nullptr;
    size_t valueLength = 0;

    ZEND_PARSE_PARAMETERS_START(2, 2)
        Z_PARAM_STRING(key, keyLength)
        Z_PARAM_STRING(value, valueLength)
    ZEND_PARSE_PARAMETERS_END();

    if (!key || !value || keyLength == 0 || valueLength == 0) {
        AIKIDO_LOG_ERROR("set_env: key or value is null!\n");
        RETURN_BOOL(false);
    }

    std::string keyStr(key, keyLength);
    std::string valueStr(value, valueLength);

    if (keyStr == "AIKIDO_TOKEN") {
        AIKIDO_GLOBAL(token) = valueStr;
    }
    else if (keyStr == "AIKIDO_DEBUG") {
        if (GetBoolFromString(valueStr, false)) {
            AIKIDO_GLOBAL(log_level_str) = "DEBUG";
            AIKIDO_GLOBAL(log_level) = AIKIDO_LOG_LEVEL_DEBUG;
        }
    }
    else if (keyStr == "AIKIDO_BLOCKING" || keyStr == "AIKIDO_BLOCK") {
        AIKIDO_GLOBAL(blocking) = GetBoolFromString(valueStr, false);
    }
    else if (keyStr == "AIKIDO_TRUST_PROXY") {
        AIKIDO_GLOBAL(trust_proxy) = GetBoolFromString(valueStr, true);
    }
    else if (keyStr == "AIKIDO_DISK_LOGS") {
        AIKIDO_GLOBAL(disk_logs) = GetBoolFromString(valueStr, false);
    }
    else if (keyStr == "AIKIDO_LOCALHOST_ALLOWED_BY_DEFAULT") {
        AIKIDO_GLOBAL(localhost_allowed_by_default) = GetBoolFromString(valueStr, true);
    }
    else if (keyStr == "AIKIDO_FEATURE_COLLECT_API_SCHEMA") {
        AIKIDO_GLOBAL(collect_api_schema) = GetBoolFromString(valueStr, true);
    }
    else {
        AIKIDO_LOG_ERROR("set_env: unknown key: %s\n", keyStr.c_str());
        RETURN_BOOL(false);
    }

    AIKIDO_LOG_INFO("set_env: %s = %s\n", keyStr.c_str(), valueStr.c_str());
    
    requestProcessor.LoadConfig(true);

    RETURN_BOOL(true);
}
