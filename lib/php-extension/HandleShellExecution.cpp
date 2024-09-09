#include "HandleShellExecution.h"
#include "Utils.h"

AIKIDO_HANDLER_FUNCTION(handle_shell_execution) {
	zend_string *cmd = NULL;

	ZEND_PARSE_PARAMETERS_START(0,-1)
		Z_PARAM_STR(cmd)
	ZEND_PARSE_PARAMETERS_END();

	std::string cmdString(ZSTR_VAL(cmd));

	std::string functionNameString(AIKIDO_GET_FUNCTION_NAME());
	
	inputEvent = {
		{ "event", "before_function_executed" },
		{ "data", {
			{ "function_name", "function_shell_executed" },
			{ "parameters", {
				{ "cmd", cmdString },
				{ "operation", functionNameString },
                { "context", get_context()}
			} }
		} }
	};
}