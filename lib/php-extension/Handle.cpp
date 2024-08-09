#include "HandleUrls.h"
#include "HandleShellExecution.h"
#include "HandlePathTraversal.h"
#include "HandlePDO.h"

#include "Utils.h"

unordered_map<std::string, PHP_HANDLERS> HOOKED_FUNCTIONS = {
	AIKIDO_REGISTER_FUNCTION_HANDLER(curl_exec),

	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(exec,                handle_shell_execution),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(shell_exec,          handle_shell_execution),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(system,              handle_shell_execution),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(passthru,            handle_shell_execution),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(popen,               handle_shell_execution),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(proc_open,           handle_shell_execution),
	
	/* Path traversal */
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(basename,             handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(chdir,             	  handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(chgrp,	              handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(chmod,                handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(chown,                handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(copy,                 handle_file_path_access_2),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(dirname,              handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(disk_free_space,      handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(disk_total_space,     handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(diskfreespace,        handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(file,                 handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(file_exists,          handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(file_get_contents,    handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(file_put_contents,    handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(fileatime,            handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(filectime,            handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(filegroup,            handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(fileinode,            handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(filemtime,            handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(fileowner,            handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(fileperms,            handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(filesize,             handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(filetype,             handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(fnmatch,              handle_file_path_access_2),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(fopen,                handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(is_dir,               handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(is_executable,        handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(is_file,              handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(is_link,              handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(is_readable,          handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(is_uploaded_file,     handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(is_writable,          handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(is_writeable,         handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(lchgrp,               handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(lchown,               handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(link,                 handle_file_path_access_2),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(linkinfo,             handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(lstat,                handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(mkdir,                handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(move_uploaded_file,   handle_file_path_access_2),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(opendir,              handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(parse_ini_file,       handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(pathinfo,             handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(readfile,             handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(readlink,             handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(realpath,             handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(rename,               handle_file_path_access_2),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(rmdir,                handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(scandir,              handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(stat,                 handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(symlink,              handle_file_path_access_2),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(touch,                handle_file_path_access),
	AIKIDO_REGISTER_FUNCTION_HANDLER_EX(unlink,               handle_file_path_access),

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