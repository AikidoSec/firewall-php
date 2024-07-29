#include "Utils.h"
#include <ctime>

std::string to_lowercase(const std::string& str) {
    std::string result = str;
    std::transform(result.begin(), result.end(), result.begin(), [](unsigned char c){ return std::tolower(c); });
    return result;
}

FILE* log_file = nullptr;

void aikido_log_init() {
    std::time_t current_time = std::time(nullptr);
    char time_str[20];
    std::strftime(time_str, sizeof(time_str), "%Y%m%d%H%M%S", std::localtime(&current_time));
    std::string log_file_path = "/var/log/aikido-" + std::string(PHP_AIKIDO_VERSION) + "/aikido-extension-php-" + time_str + ".log";
    log_file = fopen(log_file_path.c_str(), "w");
}

void aikido_log_uninit() {
    if (log_file) {
        fclose(log_file);
        log_file = nullptr;
    }
}

void aikido_log(AIKIDO_LOG_LEVEL level, const char* format, ...) {
    if (!log_file || level < AIKIDO_GLOBAL(log_level)) {
        return;
    }

    fprintf(log_file, "[AIKIDO][%s] ", aikido_log_level_str(level).c_str());

    va_list args;
    va_start(args, format);
    vfprintf(log_file, format, args);
    va_end(args);

    fflush(log_file);
}

std::string aikido_log_level_str(AIKIDO_LOG_LEVEL level) {
    switch (level) {
        case AIKIDO_LOG_LEVEL_DEBUG:
            return "DEBUG";
        case AIKIDO_LOG_LEVEL_INFO:
            return "INFO";
        case AIKIDO_LOG_LEVEL_WARN:
            return "WARN";
        case AIKIDO_LOG_LEVEL_ERROR:
            return "ERROR";
    }
    return "UNKNOWN";
}

AIKIDO_LOG_LEVEL aikido_log_level_from_str(std::string level) {
    if (level == "ERROR") {
        return AIKIDO_LOG_LEVEL_ERROR;
    }
    if (level == "WARN") {
        return AIKIDO_LOG_LEVEL_WARN;
    }
    if (level == "INFO") {
        return AIKIDO_LOG_LEVEL_INFO;
    }
    if (level == "DEBUG") {
        return AIKIDO_LOG_LEVEL_DEBUG;
    }
    return AIKIDO_LOG_LEVEL_ERROR;
}

std::string get_environment_variable(const std::string& env_key) {
    const char* env_value = getenv(env_key.c_str());
    if (!env_value) return "";
    return env_value;
}

std::string config_override_with_env(const std::string& env_key, const std::string default_value) {
	std::string env_value = get_environment_variable(env_key.c_str());
	if (!env_value.empty()) {
        return env_value;
	}
    return default_value;
}

bool config_override_with_env_bool(const std::string& env_key, bool default_value) {
	std::string env_value = get_environment_variable(env_key.c_str());
	if (!env_value.empty()) {
        return (env_value == "1" || env_value == "true");
	}
    return default_value;
}


std::string extract_server_var(zval *server, const char *var) {
    zval *data = zend_hash_str_find(Z_ARRVAL_P(server), var, strlen(var));
    if (!data) {
        return "";
    }
    return Z_STRVAL_P(data);
}

json get_route_and_method(zval *server) {
    std::string route = extract_server_var(server, "REQUEST_URI");
    std::string method = extract_server_var(server, "REQUEST_METHOD");
    // Remove query string
    size_t pos = route.find("?");
    if (pos != std::string::npos) {
        route = route.substr(0, pos);
    }
    json result = {
        {"route", route},
        {"method", method}
    };
    return result;
}

bool send_request_metadata_event(){
    zval *server = zend_hash_str_find(&EG(symbol_table), "_SERVER", sizeof("_SERVER") - 1);
    if (!server) {
        AIKIDO_LOG_WARN("\"_SERVER\" variable not found!\n");
        return false;
    }
    
    json routeAndMethod = get_route_and_method(server);
    
    if (routeAndMethod["route"].size() <= 1 || routeAndMethod["method"].size() <= 1) {
        AIKIDO_LOG_WARN("Route('%s') or method('%s') variables are empty!\n", routeAndMethod["route"].get<std::string>().c_str(), routeAndMethod["method"].get<std::string>().c_str());
        return false;
    }

    json inputEvent = {
        { "event", "request_metadata" },
        { "data", {
            { "route", routeAndMethod["route"] },
            { "method", routeAndMethod["method"] }
        }
        }
    };

    json response = GoRequestProcessorOnEvent(inputEvent);
    
    return response["status"] == "ok";
}