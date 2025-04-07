#include "Includes.h"

// Minimum value to use for reporting once every X requests the collected stats to Agent
// As the report_stats_interval_to_agent is configurable, this define is used to ensure that the configured interval is NEVER less that 50 requests
#define MIN_REPORT_STATS_INTERVAL_TO_AGENT 50

std::string GetPhpEnvVariable(const std::string& env_key) {
    zval env_value;
    if (!CallPhpFunctionWithOneParam("getenv", env_key, &env_value) || Z_TYPE(env_value) != IS_STRING) {
        return "";
    }

    std::string env_value_str = Z_STRVAL_P(&env_value);
    zval_ptr_dtor(&env_value);

    AIKIDO_LOG_DEBUG("php_env[%s] = %s\n", env_key.c_str(), env_value_str.c_str());
    return env_value_str;
} 

std::string GetSystemEnvVariable(const std::string& env_key) {
    const char* env_value = getenv(env_key.c_str());
    if (!env_value) return "";
    AIKIDO_LOG_DEBUG("sys_env[%s] = %s\n", env_key.c_str(), env_value);
    return env_value;
}

std::unordered_map<std::string, std::string> laravelEnv;
bool laravelEnvLoaded = false;

bool LoadLaravelEnvFileOnce() {
    if (laravelEnvLoaded) {
        return true;
    }

    if (!request.IsServerVarLoaded()) {
        return false;
    }
    std::string docRoot = request.GetVar("DOCUMENT_ROOT");
    if (docRoot.empty()) {
        return false;
    }
    std::string laravelEnvPath = docRoot + "/../.env";
    std::ifstream envFile(laravelEnvPath);

    if (!envFile.is_open()) {
        return false;
    }

    std::string line;
    while (std::getline(envFile, line)) {
        // Skip empty lines and comments
        if (line.empty() || line[0] == '#') {
            continue;
        }

        // Check if line starts with env_key
        if (line.substr(0, 6) == "AIKIDO") {
            size_t pos = line.find('=');
            if (pos != std::string::npos) {
                std::string key = line.substr(0, pos);
                std::string value = line.substr(pos + 1);
                
                // Trim whitespace from key and value
                key.erase(0, key.find_first_not_of(" "));
                key.erase(key.find_last_not_of(" ") + 1);
                value.erase(0, value.find_first_not_of(" "));
                value.erase(value.find_last_not_of(" ") + 1);
                
                // Remove quotes if present
                if (value.length() >= 2 && 
                    ((value.front() == '"' && value.back() == '"') ||
                     (value.front() == '\'' && value.back() == '\''))) {
                    value = value.substr(1, value.length() - 2);
                }
                laravelEnv[key] = value;
            }
        }
    }
    laravelEnvLoaded = true;
    AIKIDO_LOG_DEBUG("Loaded Laravel env file: %s\n", laravelEnvPath.c_str());
    return true;
}

std::string GetLaravelEnvVariable(const std::string& env_key) {
    LoadLaravelEnvFileOnce();
    if (laravelEnv.find(env_key) != laravelEnv.end()) {
        AIKIDO_LOG_DEBUG("laravel_env[%s] = %s\n", env_key.c_str(), laravelEnv[env_key].c_str());
        return laravelEnv[env_key];
    }
    return "";
}

/*
    Load env variables from the following sources (in this order):
    - System environment variables
    - PHP environment variables
    - Laravel environment variables
*/
using EnvGetterFn = std::string(*)(const std::string&);
EnvGetterFn envGetters[] = {
    &GetSystemEnvVariable,
    &GetPhpEnvVariable,
    &GetLaravelEnvVariable
};

std::string GetEnvVariable(const std::string& env_key) {
    for (EnvGetterFn envGetter : envGetters) {
        std::string env_value = envGetter(env_key);
        if (!env_value.empty()) {
            return env_value;
        }
    }
    return "";
}

std::string GetEnvString(const std::string& env_key, const std::string default_value) {
    std::string env_value = GetEnvVariable(env_key);
    if (!env_value.empty()) {
        return env_value;
    }
    return default_value;
}

bool GetEnvBool(const std::string& env_key, bool default_value) {
    std::string env_value = ToLowercase(GetEnvVariable(env_key));
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
            if (number <= MIN_REPORT_STATS_INTERVAL_TO_AGENT) {
                return MIN_REPORT_STATS_INTERVAL_TO_AGENT;
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

    AIKIDO_GLOBAL(blocking) = GetEnvBool("AIKIDO_BLOCK", false) || GetEnvBool("AIKIDO_BLOCKING", false);;
    AIKIDO_GLOBAL(disable) = GetEnvBool("AIKIDO_DISABLE", false);
    AIKIDO_GLOBAL(collect_api_schema) = GetEnvBool("AIKIDO_FEATURE_COLLECT_API_SCHEMA", true);
    AIKIDO_GLOBAL(localhost_allowed_by_default) = GetEnvBool("AIKIDO_LOCALHOST_ALLOWED_BY_DEFAULT", true);
    AIKIDO_GLOBAL(trust_proxy) = GetEnvBool("AIKIDO_TRUST_PROXY", true);
    AIKIDO_GLOBAL(disk_logs) = GetEnvBool("AIKIDO_DISK_LOGS", false);
    AIKIDO_GLOBAL(sapi_name) = sapi_module.name;
    AIKIDO_GLOBAL(token) = GetEnvString("AIKIDO_TOKEN", "");
    AIKIDO_GLOBAL(endpoint) = GetEnvString("AIKIDO_ENDPOINT", "https://guard.aikido.dev/");
    AIKIDO_GLOBAL(config_endpoint) = GetEnvString("AIKIDO_REALTIME_ENDPOINT", "https://runtime.aikido.dev/");
    AIKIDO_GLOBAL(report_stats_interval_to_agent) = GetEnvNumber("AIKIDO_REPORT_STATS_INTERVAL", 10000);
}