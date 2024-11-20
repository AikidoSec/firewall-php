#include "Includes.h"

#define MIN_REPORT_STATS_INTERVAL 50

std::string GetEnvVariable(const std::string& env_key) {
    const char* env_value = getenv(env_key.c_str());
    if (!env_value) return "";
    AIKIDO_LOG_DEBUG("env[%s] = %s\n", env_key.c_str(), env_value);
    return env_value;
}

std::string GetEnvString(const std::string& env_key, const std::string default_value) {
    std::string env_value = GetEnvVariable(env_key.c_str());
    if (!env_value.empty()) {
        return env_value;
    }
    return default_value;
}

bool GetEnvBool(const std::string& env_key, bool default_value) {
    std::string env_value = GetEnvVariable(env_key.c_str());
    if (!env_value.empty()) {
        return (env_value == "1" || env_value == "true");
    }
    return default_value;
}

unsigned int GetEnvNumber(const std::string& env_key, unsigned int default_value) {
    std::string env_value = GetEnvVariable(env_key.c_str());
    if (!env_value.empty()) {
        try {
            unsigned int number = std::stoi(env_value);
            if (number <= MIN_REPORT_STATS_INTERVAL) {
                return MIN_REPORT_STATS_INTERVAL;
            }
        }
        catch (...) {}
    }
    return default_value;
}

void LoadEnvironment() {
    if (GetEnvBool("AIKIDO_DEBUG", false)) {
        AIKIDO_GLOBAL(log_level_str) = "DEBUG";
        AIKIDO_GLOBAL(log_level) = AIKIDO_LOG_LEVEL_DEBUG;
    } else {
        AIKIDO_GLOBAL(log_level_str) = GetEnvString("AIKIDO_LOG_LEVEL", "WARN");
        AIKIDO_GLOBAL(log_level) = Log::ToLevel(AIKIDO_GLOBAL(log_level_str));
    }

    AIKIDO_GLOBAL(blocking) = GetEnvBool("AIKIDO_BLOCK", false);
    AIKIDO_GLOBAL(disable) = GetEnvBool("AIKIDO_DISABLE", false);
    AIKIDO_GLOBAL(collect_api_schema) = GetEnvBool("AIKIDO_FEATURE_COLLECT_API_SCHEMA", true);
    AIKIDO_GLOBAL(localhost_allowed_by_default) = GetEnvBool("AIKIDO_LOCALHOST_ALLOWED_BY_DEFAULT", true);
    AIKIDO_GLOBAL(trust_proxy) = GetEnvBool("AIKIDO_TRUST_PROXY", true);
    AIKIDO_GLOBAL(socket_path) = GenerateSocketPath();
    AIKIDO_GLOBAL(sapi_name) = sapi_module.name;
    AIKIDO_GLOBAL(token) = GetEnvString("AIKIDO_TOKEN", "");
    AIKIDO_GLOBAL(endpoint) = GetEnvString("AIKIDO_ENDPOINT", "https://guard.aikido.dev/");
    AIKIDO_GLOBAL(config_endpoint) = GetEnvString("AIKIDO_REALTIME_ENDPOINT", "https://runtime.aikido.dev/");
    AIKIDO_GLOBAL(report_stats_interval) = GetEnvNumber("AIKIDO_REPORT_STATS_INTERVAL", 10000);
}