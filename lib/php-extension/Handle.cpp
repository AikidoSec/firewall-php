#include "HandleUrls.h"
#include "HandleShellExecution.h"
#include "HandlePDO.h"

#include "Utils.h"

unordered_map<std::string, PHP_HANDLERS> HOOKED_FUNCTIONS = {
	AIKIDO_REGISTER_FUNCTION_HANDLER(curl_exec),

	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(exec,       handle_shell_execution),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(shell_exec, handle_shell_execution),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(system,     handle_shell_execution),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(passthru,   handle_shell_execution),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(popen,      handle_shell_execution),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(proc_open,  handle_shell_execution)
};

unordered_map<AIKIDO_METHOD_KEY, PHP_HANDLERS, AIKIDO_METHOD_KEY_HASH> HOOKED_METHODS = {
	AIKIDO_REGISTER_METHOD_HANDLER(pdo, __construct),
	AIKIDO_REGISTER_METHOD_HANDLER(pdo, query)
};

ZEND_NAMED_FUNCTION(aikido_generic_handler) {
	AIKIDO_LOG_DEBUG("Aikido generic handler started!\n");

	zif_handler original_handler = nullptr;

	if (request_processor_on_event_fn) {
		try {
			zend_execute_data *exec_data = EG(current_execute_data);
			zend_function *func = exec_data->func;
			zend_class_entry* executed_scope = zend_get_executed_scope();
			
			std::string function_name(ZSTR_VAL(func->common.function_name));
			function_name = to_lowercase(function_name);

			aikido_handler handler = nullptr;

			std::string scope_name = function_name;
			AIKIDO_LOG_DEBUG("Function name: %s\n", scope_name.c_str());
			if (HOOKED_FUNCTIONS.find(function_name) != HOOKED_FUNCTIONS.end()) {
				handler = HOOKED_FUNCTIONS[function_name].handler;
				original_handler = HOOKED_FUNCTIONS[function_name].original_handler;
			}
			else if (executed_scope) {
				/* A method was executed (executed_scope stores the name of the current class) */

				std::string class_name(ZSTR_VAL(executed_scope->name));
				class_name = to_lowercase(class_name);

				scope_name = class_name + "->" + function_name;

				AIKIDO_METHOD_KEY method_key(class_name, function_name);

				AIKIDO_LOG_DEBUG("Method name: %s\n", scope_name.c_str());

				if (HOOKED_METHODS.find(method_key) == HOOKED_METHODS.end()) {
					AIKIDO_LOG_DEBUG("Method not found! Returning!\n");
					return;
				}

				handler = HOOKED_METHODS[method_key].handler;
				original_handler = HOOKED_METHODS[method_key].original_handler;
			}
			else {
				AIKIDO_LOG_DEBUG("Nothing matches the current handler! Returning!\n");
				return;
			}

			AIKIDO_LOG_DEBUG("Calling handler for \"%s\"!\n", scope_name.c_str());

			json inputEvent;
			handler(INTERNAL_FUNCTION_PARAM_PASSTHRU, inputEvent);

			if (!inputEvent.empty()) {
				json outputEvent = GoRequestProcessorOnEvent(inputEvent);
				if (AIKIDO_GLOBAL(blocking) == true && aikido_execute_output(outputEvent) == BLOCK) {
					// exit generic handler and do not call the original handler
					// thus blocking the execution 
					return;
				}
			}
		}
		catch (const std::exception& e) {
			AIKIDO_LOG_ERROR("Exception encountered in generic handler: %s\n", e.what());
		}
	}
	
	if (original_handler) {
		original_handler(INTERNAL_FUNCTION_PARAM_PASSTHRU);
	}

	AIKIDO_LOG_DEBUG("Aikido generic handler ended!\n");
}