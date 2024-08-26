#pragma once

#include "Includes.h"

/* 
	Macro for registering an Aikido handler in the HOOKED_FUNCTIONS map.
	It takes as parameters the PHP function name to be hooked and C++ function 
		that should be called when that PHP function is executed.
	The nullptr part is a placeholder where the original function handler from
		the Zend framework will be stored at initialization when we run the hooking.
*/
#define AIKIDO_REGISTER_FUNCTION_HANDLER_EX(function_name, function_pointer) { std::string(#function_name), { function_pointer, nullptr, nullptr } }

/*
	Shorthand version of AIKIDO_REGISTER_FUNCTION_HANDLER_EX that constructs automatically the C++ function to be called.
	This version only registers a pre-hook (hook to be called before the original function is executed).
	For example, if function name is curl_init this macro will store { "curl_init", { handle_pre_curl_init, nullptr } }.
*/
#define AIKIDO_REGISTER_FUNCTION_HANDLER(function_name) { std::string(#function_name), { handle_pre_##function_name, nullptr, nullptr } }

/*
	Shorthand version of AIKIDO_REGISTER_FUNCTION_HANDLER_EX that constructs automatically the C++ function to be called.
	This version registers a pre-hook and a post-hook (hooks for before and after the function is executed).
	For example, if function name is curl_init this macro will store { "curl_init", { handle_pre_curl_init, handle_post_curl_init, nullptr } }.
*/
#define AIKIDO_REGISTER_FUNCTION_HANDLER_WITH_POST(function_name) { std::string(#function_name), { handle_pre_##function_name, handle_post_##function_name, nullptr } }


/*
	Similar to AIKIDO_REGISTER_FUNCTION_HANDLER, but for methods.
*/
#define AIKIDO_REGISTER_METHOD_HANDLER(class_name, method_name) { AIKIDO_METHOD_KEY(std::string(#class_name), std::string(#method_name)), { handle_pre_ ## class_name ## _ ## method_name, nullptr } }

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

std::string aikido_log_level_str(AIKIDO_LOG_LEVEL level);

AIKIDO_LOG_LEVEL aikido_log_level_from_str(std::string level);

std::string to_lowercase(const std::string& str);

std::string get_environment_variable(const std::string& env_key);

std::string get_env_string(const std::string& env_key, const std::string default_value);

bool get_env_bool(const std::string& env_key, bool default_value);

enum ACTION {
	CONTINUE,
	BLOCK,
	EXIT
};

ACTION send_request_init_metadata_event();
ACTION send_request_shutdown_metadata_event();

ACTION aikido_execute_output(json event);

json get_context();

bool send_user_event(std::string id, std::string username);

bool aikido_call_user_function(std::string function_name, unsigned int params_number = 0, zval* params = nullptr);

bool aikido_call_user_function_one_param(std::string function_name, long first_param);

bool aikido_call_user_function_one_param(std::string function_name, std::string first_param);
