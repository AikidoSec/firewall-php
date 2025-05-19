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
