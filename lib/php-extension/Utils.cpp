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

std::string get_env_string(const std::string& env_key, const std::string default_value) {
	std::string env_value = get_environment_variable(env_key.c_str());
	if (!env_value.empty()) {
        return env_value;
	}
    return default_value;
}

bool get_env_bool(const std::string& env_key, bool default_value) {
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


int get_status_code() {
    int status_code = SG(sapi_headers).http_response_code;
    return status_code;
}


std::string extract_body() {
    long maxlen = PHP_STREAM_COPY_ALL;
    zend_string *contents;
    php_stream *stream;

    stream = php_stream_open_wrapper("php://input", "rb", 0 | REPORT_ERRORS, NULL);
    if ((contents = php_stream_copy_to_mem(stream, maxlen, 0)) != NULL) {
        php_stream_close(stream);
        return std::string(ZSTR_VAL(contents));
    }
    php_stream_close(stream);
    return "";
}

json extract_headers(zval *server) {
    std::map<std::string, std::string> headers;
    zend_string *key;
    zval *val;
    ZEND_HASH_FOREACH_STR_KEY_VAL(Z_ARRVAL_P(server), key, val) {
        if (key && ZSTR_LEN(key) > 5 && memcmp(ZSTR_VAL(key), "HTTP_", 5) == 0) {
            std::string header(ZSTR_VAL(key) + 5);
            std::transform(header.begin(), header.end(), header.begin(), ::tolower);
            headers[header] = Z_STRVAL_P(val);
        }
    } ZEND_HASH_FOREACH_END();

  
    json headers_json;
    for (auto const& [key, val] : headers) {
        headers_json[key] = val;
    }
    return headers_json;
}

json get_context() {
    zval *server = zend_hash_str_find(&EG(symbol_table), "_SERVER", sizeof("_SERVER") - 1);
    bool https = extract_server_var(server, "HTTPS") != "" ? true : false;

    if (server && Z_TYPE_P(server) == IS_ARRAY) {
        return {
            { "https", https },
            { "url", (https ? "https://" : "http://") + extract_server_var(server, "HTTP_HOST") + extract_server_var(server, "REQUEST_URI") },
            { "method", extract_server_var(server, "REQUEST_METHOD") },
            { "query", extract_server_var(server, "QUERY_STRING") },
            { "headers", extract_headers(server) },
            { "remoteAddress", extract_server_var(server, "REMOTE_ADDR") },
            { "cookies", extract_server_var(server, "HTTP_COOKIE") },
            { "body", extract_body() }
        };
    }

    return {};
}

ACTION send_request_init_metadata_event(){
    zval *server = zend_hash_str_find(&EG(symbol_table), "_SERVER", sizeof("_SERVER") - 1);
    if (!server) {
        AIKIDO_LOG_WARN("\"_SERVER\" variable not found!\n");
        return CONTINUE;
    }
    
    json routeAndMethod = get_route_and_method(server);

    json inputEvent = {
        { "event", "request_init" },
        { "data", {
            { "route", routeAndMethod["route"] },
            { "method", routeAndMethod["method"] },
            { "remoteAddress", extract_server_var(server, "REMOTE_ADDR") },
            { "xForwardedFor",  extract_server_var(server, "HTTP_X_FORWARDED_FOR") },
        }
        }
    };

    try {
        json response = GoRequestProcessorOnEvent(inputEvent);
        return aikido_execute_output(response);
    }
    catch (const std::exception& e) {
        AIKIDO_LOG_ERROR("Exception encountered in processing request metadata: %s\n", e.what());
    }
    return CONTINUE;
}

ACTION send_request_shutdown_metadata_event(){
    zval *server = zend_hash_str_find(&EG(symbol_table), "_SERVER", sizeof("_SERVER") - 1);
    if (!server) {
        AIKIDO_LOG_WARN("\"_SERVER\" variable not found!\n");
        return CONTINUE;
    }
    
    json routeAndMethod = get_route_and_method(server);

    json inputEvent = {
        { "event", "request_shutdown" },
        { "data", {
            { "route", routeAndMethod["route"] },
            { "method", routeAndMethod["method"] },
            { "status_code", get_status_code() }
        }
        }
    };

    try {
        json response = GoRequestProcessorOnEvent(inputEvent);
        return aikido_execute_output(response);
    }
    catch (const std::exception& e) {
        AIKIDO_LOG_ERROR("Exception encountered in processing request metadata: %s\n", e.what());
    }
    return CONTINUE;
}


bool send_user_event(std::string id, std::string username) {
    zval *server = zend_hash_str_find(&EG(symbol_table), "_SERVER", sizeof("_SERVER") - 1);
    if (!server) {
        AIKIDO_LOG_WARN("\"_SERVER\" variable not found!\n");
        return false;
    }

    json inputEvent = {
        { "event", "user_event" },
        { "data", { 
            { "id", id },
            { "username", username },
            { "remoteAddress", extract_server_var(server, "REMOTE_ADDR") },
            { "xForwardedFor",  extract_server_var(server, "HTTP_X_FORWARDED_FOR") }
        }
        }
    };

    try {
        json response = GoRequestProcessorOnEvent(inputEvent);
        aikido_execute_output(response);
        return true;
    }
    catch (const std::exception& e) {
        AIKIDO_LOG_ERROR("Exception encountered in processing user event: %s\n", e.what());
    }
    return false;
}

bool aikido_call_user_function(std::string function_name, unsigned int params_number, zval* params) {
    zval _function_name;
    zend_string* _function_name_str = zend_string_init(function_name.c_str(), function_name.length(), 0);
    ZVAL_STR(&_function_name, _function_name_str);

    zval _return_value;

    int _result = call_user_function(EG(function_table), nullptr, &_function_name, &_return_value, params_number, params);

    zend_string_release(_function_name_str);
    zval_ptr_dtor(&_return_value);

    return _result == SUCCESS;
}

bool aikido_call_user_function_one_param(std::string function_name, long first_param) {
    zval _params[1];
    ZVAL_LONG(&_params[0], first_param);
    return aikido_call_user_function(function_name, 1, _params);
}

bool aikido_call_user_function_one_param(std::string function_name, std::string first_param) {
    zval _params[1];
    zend_string* _first_param = zend_string_init(first_param.c_str(), first_param.length(), 0);
    ZVAL_STR(&_params[0], _first_param);

    bool ret = aikido_call_user_function(function_name, 1, _params);

    zend_string_release(_first_param);

    return ret;
}
