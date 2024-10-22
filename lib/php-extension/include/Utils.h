#pragma once

#include "Includes.h"

/*
	Macro for registering an Aikido handler in the HOOKED_FUNCTIONS map.
	It takes as parameters the PHP function name to be hooked, a C++ function
	that should be called BEFORE that PHP function is executed and a C++ function
	that should be called AFTER that PHP function is executed.
	The nullptr part is a placeholder where the original function handler from
	the Zend framework will be stored at initialization when we run the hooking.
*/
#define AIKIDO_REGISTER_FUNCTION_HANDLER_WITH_POST_EX(function_name, pre_handler, post_handler) { std::string(#function_name), { pre_handler, post_handler, nullptr } }

/* 
	Macro for registering an Aikido handler in the HOOKED_FUNCTIONS map.
	It takes as parameters the PHP function name to be hooked and a C++ function 
	that should be called BEFORE that PHP function is executed.
	This macro doesn't register any hook for AFTER the function is execute, that's why
	the second argument is nullptr.
	The last nullptr part is a placeholder where the original function handler from
	the Zend framework will be stored at initialization when we run the hooking.
*/
#define AIKIDO_REGISTER_FUNCTION_HANDLER_EX(function_name, pre_handler) AIKIDO_REGISTER_FUNCTION_HANDLER_WITH_POST_EX(function_name, pre_handler, nullptr)

/*
	Shorthand version of AIKIDO_REGISTER_FUNCTION_HANDLER_EX that constructs automatically the C++ function to be called.
	This version only registers a pre-hook (hook to be called before the original function is executed).
	For example, if function name is curl_init this macro will store { "curl_init", { handle_pre_curl_init, nullptr } }.
*/
#define AIKIDO_REGISTER_FUNCTION_HANDLER(function_name) { std::string(#function_name), { handle_pre_##function_name, nullptr, nullptr } }

/*
	Shorthand version of AIKIDO_REGISTER_FUNCTION_HANDLER_WITH_POST_EX that constructs automatically the C++ function to be called.
	This version registers a pre-hook and a post-hook (hooks for before and after the function is executed).
	For example, if function name is curl_init this macro will store { "curl_init", { handle_pre_curl_init, handle_post_curl_init, nullptr } }.
*/
#define AIKIDO_REGISTER_FUNCTION_HANDLER_WITH_POST(function_name) AIKIDO_REGISTER_FUNCTION_HANDLER_WITH_POST_EX(function_name, handle_pre_##function_name, handle_post_##function_name)

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

enum ACTION_STATUS {
	CONTINUE,
	BLOCK,
	EXIT
};

void send_request_init_metadata_event();

void send_request_shutdown_metadata_event();

std::string extract_server_var(const char *var);

std::string extract_body();

std::string extract_route();

std::string extract_status_code();

std::string extract_url();

std::string extract_headers();

bool aikido_echo(std::string s);

bool aikido_call_user_function(std::string function_name, unsigned int params_number = 0, zval *params = nullptr, zval *return_value = nullptr, zval *object = nullptr);

bool aikido_call_user_function_one_param(std::string function_name, long first_param, zval *return_value = nullptr, zval *object = nullptr);

bool aikido_call_user_function_one_param(std::string function_name, std::string first_param, zval *return_value = nullptr, zval *object = nullptr);

std::string aikido_call_user_function_curl_getinfo(zval *curl_handle, int curl_info_option);

std::string aikido_generate_socket_path();

const char *get_event_name(EVENT_ID event);
