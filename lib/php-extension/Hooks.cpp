#include "Includes.h"

unordered_map<std::string, PHP_HANDLERS> HOOKED_FUNCTIONS = {
    /* Outgoing request */
    AIKIDO_REGISTER_FUNCTION_HANDLER_WITH_POST(curl_exec),

    /* Shell execution */
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(exec, handle_shell_execution),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(shell_exec, handle_shell_execution),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(system, handle_shell_execution),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(passthru, handle_shell_execution),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(popen, handle_shell_execution),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(proc_open, handle_shell_execution),

    /* Path access */
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(chdir, handle_pre_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(chgrp, handle_pre_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(chmod, handle_pre_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(chown, handle_pre_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(copy, handle_pre_file_path_access_2),
    AIKIDO_REGISTER_FUNCTION_HANDLER_WITH_POST_EX(file, handle_pre_file_path_access, handle_post_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_WITH_POST_EX(file_get_contents, handle_pre_file_path_access, handle_post_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(file_put_contents, handle_pre_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_WITH_POST_EX(fopen, handle_pre_file_path_access, handle_post_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(lchgrp, handle_pre_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(lchown, handle_pre_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(link, handle_pre_file_path_access_2),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(mkdir, handle_pre_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(move_uploaded_file, handle_pre_file_path_access_2),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(opendir, handle_pre_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(parse_ini_file, handle_pre_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(readfile, handle_pre_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(readlink, handle_pre_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(rename, handle_pre_file_path_access_2),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(rmdir, handle_pre_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(scandir, handle_pre_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(symlink, handle_pre_file_path_access_2),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(touch, handle_pre_file_path_access),
    AIKIDO_REGISTER_FUNCTION_HANDLER_EX(unlink, handle_pre_file_path_access),
};

unordered_map<AIKIDO_METHOD_KEY, PHP_HANDLERS, AIKIDO_METHOD_KEY_HASH> HOOKED_METHODS = {
    AIKIDO_REGISTER_METHOD_HANDLER(pdo, query)};

void HookFunctions() {
    for (auto &it : HOOKED_FUNCTIONS) {
        zend_function *function_data = (zend_function *)zend_hash_str_find_ptr(CG(function_table), it.first.c_str(), it.first.length());
        if (function_data == NULL) {
            AIKIDO_LOG_WARN("Function \"%s\" does not exist!\n", it.first.c_str());
            continue;
        }
        if (it.second.original_handler) {
            AIKIDO_LOG_WARN("Function \"%s\" already hooked (original handler %p)!\n", it.first.c_str(), it.second.original_handler);
            continue;
        }

        it.second.original_handler = function_data->internal_function.handler;
        function_data->internal_function.handler = aikido_generic_handler;
        AIKIDO_LOG_INFO("Hooked function \"%s\" (original handler %p)!\n", it.first.c_str(), it.second.original_handler);
    }
}

void HookMethods() {
    for (auto &it : HOOKED_METHODS) {
        zend_class_entry *class_entry = (zend_class_entry *)zend_hash_str_find_ptr(CG(class_table), it.first.class_name.c_str(), it.first.class_name.length());
        if (class_entry == NULL) {
            AIKIDO_LOG_WARN("Class \"%s\" does not exist!\n", it.first.class_name.c_str());
            continue;
        }

        zend_function *method = (zend_function *)zend_hash_str_find_ptr(&class_entry->function_table, it.first.method_name.c_str(), it.first.method_name.length());
        if (method == NULL) {
            AIKIDO_LOG_WARN("Method \"%s->%s\" does not exist!\n", it.first.class_name.c_str(), it.first.method_name.c_str());
            continue;
        }

        if (it.second.original_handler) {
            AIKIDO_LOG_WARN("Method \"%s->%s\" already hooked (original handler %p)!\n", it.first.class_name.c_str(), it.first.method_name.c_str(), it.second.original_handler);
            continue;
        }

        it.second.original_handler = method->internal_function.handler;
        method->internal_function.handler = aikido_generic_handler;
        AIKIDO_LOG_INFO("Hooked method \"%s->%s\" (original handler %p)!\n", it.first.class_name.c_str(), it.first.method_name.c_str(), it.second.original_handler);
    }
}

static void (*original_zend_execute_ex)(zend_execute_data *execute_data) = NULL;

void aikido_zend_execute_ex(zend_execute_data *execute_data) {
    if (action.Exit()) {
        AIKIDO_LOG_INFO("Current request is marked for exit. Bailing out...\n");
        zend_bailout();
    }
    original_zend_execute_ex(execute_data);
}

void HookExecute() {
    original_zend_execute_ex = zend_execute_ex;
    zend_execute_ex = aikido_zend_execute_ex;
}