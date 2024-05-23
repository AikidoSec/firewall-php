/* Aikido runtime extension for PHP */

#ifdef HAVE_CONFIG_H
# include "config.h"
#endif

#include "php.h"
#include "ext/standard/info.h"
#include "php_aikido.h"
#include <unordered_map>
#include <set>
#include <string>
#include <curl/curl.h>
#include "libaikido_go.h"
#include "GoWrappers.h"
#include "3rdparty/json.hpp"

using namespace std;
using json = nlohmann::json;

ZEND_NAMED_FUNCTION(handle_curl_init);
ZEND_NAMED_FUNCTION(handle_curl_setopt);
ZEND_NAMED_FUNCTION(handle_shell_execution);


struct FUNCTION_HANDLERS {
	zif_handler aikido_handler;
	zif_handler original_handler;
};

#define AIKIDO_REGISTER_HANDLER(function_name) { std::string(#function_name), { handle_##function_name, nullptr } }
#define AIKIDO_REGISTER_HANDLER_EX(function_name, function_pointer) { std::string(#function_name), { function_pointer, nullptr } }


unordered_map<std::string, FUNCTION_HANDLERS> HOOKED_FUNCTIONS = {
	AIKIDO_REGISTER_HANDLER(curl_init),
	AIKIDO_REGISTER_HANDLER(curl_setopt),

	AIKIDO_REGISTER_HANDLER_EX(exec,       handle_shell_execution),
	AIKIDO_REGISTER_HANDLER_EX(shell_exec, handle_shell_execution),
	AIKIDO_REGISTER_HANDLER_EX(system,     handle_shell_execution),
	AIKIDO_REGISTER_HANDLER_EX(passthru,   handle_shell_execution),
	AIKIDO_REGISTER_HANDLER_EX(popen,      handle_shell_execution),
	AIKIDO_REGISTER_HANDLER_EX(proc_open,  handle_shell_execution)
};

#define AIKIDO_HANDLER_START(function_name) php_printf("[AIKIDO-C++] Handler called for \"" #function_name "\"!\n");
#define AIKIDO_HANDLER_END(function_name) HOOKED_FUNCTIONS[#function_name].original_handler(INTERNAL_FUNCTION_PARAM_PASSTHRU);

#define AIKIDO_GET_FUNCTION_NAME() (ZSTR_VAL(execute_data->func->common.function_name))

#define AIKIDO_HANDLER_START_GENERIC() php_printf("[AIKIDO-C++] Handler called for \"%s\"!\n", AIKIDO_GET_FUNCTION_NAME());
#define AIKIDO_HANDLER_END_GENERIC() HOOKED_FUNCTIONS[AIKIDO_GET_FUNCTION_NAME()].original_handler(INTERNAL_FUNCTION_PARAM_PASSTHRU);

ZEND_NAMED_FUNCTION(handle_curl_init) {
	AIKIDO_HANDLER_START(curl_init);

	zend_string *url = NULL;

	ZEND_PARSE_PARAMETERS_START(0,1)
		Z_PARAM_OPTIONAL
		Z_PARAM_STR_OR_NULL(url)
	ZEND_PARSE_PARAMETERS_END();

	AIKIDO_HANDLER_END(curl_init);
	
	if (Z_TYPE_P(return_value) != IS_FALSE) {
		// Z_OBJ_P(return_value)
		json curl_init_event = {
			{ "event", "function_executed" },
			{ "data", {
				{ "function_name", "curl_init" },
				{ "parameters", json::object() }
			} }
		};
		if (url) {
			std::string urlString(ZSTR_VAL(url));
			curl_init_event["data"]["parameters"]["url"] = urlString;
		}
		GoOnEvent(curl_init_event);
	}
}

ZEND_NAMED_FUNCTION(handle_curl_setopt) {
	AIKIDO_HANDLER_START(curl_setopt);

	zval *curlHandle = NULL;
	zend_long options = 0;
	zval *zvalue = NULL;

	ZEND_PARSE_PARAMETERS_START(3, 3)
		Z_PARAM_OBJECT(curlHandle)
		Z_PARAM_LONG(options)
		Z_PARAM_ZVAL(zvalue)
	ZEND_PARSE_PARAMETERS_END();

	if (options == CURLOPT_URL) {
		zend_string *tmp_str;
		zend_string *url = zval_get_tmp_string(zvalue, &tmp_str);

		std::string urlString(ZSTR_VAL(url));
	
		json curl_setopt_event = {
			{ "event", "function_executed" },
			{ "data", {
				{ "function_name", "curl_setopt" },
				{ "parameters", {
					{ "url", urlString }
				} }
			} }
		};

		GoOnEvent(curl_setopt_event);

		zend_tmp_string_release(tmp_str);
	}

	AIKIDO_HANDLER_END(curl_setopt);
}

ZEND_NAMED_FUNCTION(handle_shell_execution) {
	AIKIDO_HANDLER_START_GENERIC();

	zend_string *cmd = NULL;

	ZEND_PARSE_PARAMETERS_START(1,-1)
		Z_PARAM_OPTIONAL
		Z_PARAM_STR(cmd)
	ZEND_PARSE_PARAMETERS_END();

	std::string cmdString(ZSTR_VAL(cmd));

	std::string functionNameString(AIKIDO_GET_FUNCTION_NAME());
	
	json shell_execution_event = {
		{ "event", "function_executed" },
		{ "data", {
			{ "function_name", functionNameString },
			{ "parameters", {
				{ "cmd", cmdString }
			} }
		} }
	};

	GoOnEvent(shell_execution_event);

	AIKIDO_HANDLER_END_GENERIC();
}

/* For compatibility with older PHP versions */
#ifndef ZEND_PARSE_PARAMETERS_NONE
#define ZEND_PARSE_PARAMETERS_NONE() \
	ZEND_PARSE_PARAMETERS_START(0, 0) \
	ZEND_PARSE_PARAMETERS_END()
#endif

PHP_MINIT_FUNCTION(aikido)
{
#if defined(COMPILE_DL_MY_EXTENSION) && defined(ZTS)
	ZEND_TSRMLS_CACHE_UPDATE();
#endif

	for ( auto& it : HOOKED_FUNCTIONS ) {
		zend_function* function_data = (zend_function*)zend_hash_str_find_ptr(CG(function_table), it.first.c_str(), it.first.length());
		if (function_data != NULL) {
			it.second.original_handler = function_data->internal_function.handler;
			function_data->internal_function.handler = it.second.aikido_handler;
			php_printf("[AIKIDO-C++] Hooked function \"%s\" using aikido handler %p (original handler %p)!\n", it.first.c_str(), it.second.aikido_handler, it.second.original_handler);
		}
	}

	return SUCCESS;
}

PHP_MSHUTDOWN_FUNCTION(aikido)
{
	return SUCCESS;
}

PHP_MINFO_FUNCTION(aikido)
{
	php_info_print_table_start();
	php_info_print_table_row(2, "aikido support", "enabled");
	php_info_print_table_end();
}

static const zend_function_entry ext_functions[] = {
	ZEND_FE_END
};

zend_module_entry aikido_module_entry = {
	STANDARD_MODULE_HEADER,
	"aikido",					/* Extension name */
	ext_functions,				/* zend_function_entry */
	PHP_MINIT(aikido),			/* PHP_MINIT - Module initialization */
	PHP_MSHUTDOWN(aikido),		/* PHP_MSHUTDOWN - Module shutdown */
	NULL,						/* PHP_RINIT - Request initialization */
	NULL,						/* PHP_RSHUTDOWN - Request shutdown */
	PHP_MINFO(aikido),			/* PHP_MINFO - Module info */
	PHP_AIKIDO_VERSION,			/* Version */
	STANDARD_MODULE_PROPERTIES
};

#ifdef COMPILE_DL_AIKIDO
# ifdef ZTS
ZEND_TSRMLS_CACHE_DEFINE()
# endif
ZEND_GET_MODULE(aikido)
#endif
