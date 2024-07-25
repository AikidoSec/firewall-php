#pragma once

#include "Includes.h"

/* 
	Macro for registering an Aikido handler in the HOOKED_FUNCTIONS map.
	It takes as parameters the PHP function name to be hooked and C++ function 
		that should be called when that PHP function is executed.
	The nullptr part is a placeholder where the original function handler from
		the Zend framework will be stored at initialization when we run the hooking.
*/
#define AIKIDO_REGISTER_FUNCTION_HANDLER_EX(function_name, function_pointer) { std::string(#function_name), { function_pointer, nullptr } }

/*
	Shorthand version of AIKIDO_REGISTER_FUNCTION_HANDLER_EX that constructs automatically the C++ function to be called.
	For example, if function name is curl_init this macro will store { "curl_init", { handle_curl_init, nullptr } }.
*/
#define AIKIDO_REGISTER_FUNCTION_HANDLER(function_name) { std::string(#function_name), { handle_##function_name, nullptr } }

/*
	Similar to AIKIDO_REGISTER_FUNCTION_HANDLER, but for methods.
*/
#define AIKIDO_REGISTER_METHOD_HANDLER(class_name, method_name) { AIKIDO_METHOD_KEY(std::string(#class_name), std::string(#method_name)), { handle_ ## class_name ## _ ## method_name, nullptr } }

#define AIKIDO_GET_FUNCTION_NAME() (ZSTR_VAL(execute_data->func->common.function_name))

enum AIKIDO_LOG_LEVEL {
	AIKIDO_LOG_LEVEL_DEBUG,
	AIKIDO_LOG_LEVEL_INFO,
	AIKIDO_LOG_LEVEL_WARN,
    AIKIDO_LOG_LEVEL_ERROR
};

void aikido_log_init();

void aikido_log_uninit();

void aikido_log(AIKIDO_LOG_LEVEL level, const char* format, ...);


#if defined(ZEND_DEBUG)
	#define AIKIDO_LOG_DEBUG(format, ...)  aikido_log(AIKIDO_LOG_LEVEL_DEBUG, format, ##__VA_ARGS__)
#else
	/* Disable debugging logs for production builds */
	#define AIKIDO_LOG_DEBUG(format, ...)
#endif

#define AIKIDO_LOG_INFO(format, ...)   aikido_log(AIKIDO_LOG_LEVEL_INFO, format, ##__VA_ARGS__)
#define AIKIDO_LOG_WARN(format, ...)   aikido_log(AIKIDO_LOG_LEVEL_WARN, format, ##__VA_ARGS__)
#define AIKIDO_LOG_ERROR(format, ...)  aikido_log(AIKIDO_LOG_LEVEL_ERROR, format, ##__VA_ARGS__)

const char* aikido_log_level_str(AIKIDO_LOG_LEVEL level);

AIKIDO_LOG_LEVEL aikido_log_level_from_str(std::string level);

std::string to_lowercase(const std::string& str);

std::string get_environment_variable(const std::string& env_key);

std::string config_override_with_env(const std::string& env_key, const std::string default_value);

bool config_override_with_env_bool(const std::string& env_key, bool default_value);
