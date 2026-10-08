#include "Includes.h"

/*
    This is a custom destructor, used for cleaning the allocated ast hashtable.
    This is needed because the ast hashtable is not cleaned by the zend_ast_process function.
*/
void ast_to_clean_dtor(zval *zv) {
    zend_ast *ast = (zend_ast *)Z_PTR_P(zv);
    efree(ast);
} 

void ensure_ast_hashtable_initialized() {
    auto& global_ast_to_clean = AIKIDO_GLOBAL(globalAstToClean);
    if (!global_ast_to_clean) {
        ALLOC_HASHTABLE(global_ast_to_clean);
        zend_hash_init(global_ast_to_clean, 8, NULL, ast_to_clean_dtor, 0);
    }
}

/*
    Transform ZEND_AST_SHELL_EXEC (backtick operator) into a shell_exec() function call.
    This ensures backtick expressions go through the same security hooks as shell_exec().
*/
zend_ast* transform_shell_exec_to_call(zend_ast *shell_exec_ast) {
    ensure_ast_hashtable_initialized();
    auto& global_ast_to_clean = AIKIDO_GLOBAL(globalAstToClean);

    AIKIDO_LOG_DEBUG("Transforming backtick shell execution to shell_exec() call at line %d\n", shell_exec_ast->lineno);

    // Create function name node for "shell_exec"
    zend_ast_zval *name_var = (zend_ast_zval*)emalloc(sizeof(zend_ast_zval));
    name_var->kind = ZEND_AST_ZVAL;
    ZVAL_STRING(&name_var->val, "shell_exec");
    name_var->val.u2.lineno = shell_exec_ast->lineno;
    zend_hash_next_index_insert_ptr(global_ast_to_clean, name_var);

    // Create argument list with the shell command as the first argument
    zend_ast_list *arg_list = (zend_ast_list*)emalloc(sizeof(zend_ast_list) + sizeof(zend_ast*));
    arg_list->kind = ZEND_AST_ARG_LIST;
    arg_list->lineno = shell_exec_ast->lineno;
    arg_list->children = 1;
    arg_list->child[0] = shell_exec_ast->child[0]; // The command expression from backticks
    zend_hash_next_index_insert_ptr(global_ast_to_clean, arg_list);

    // Create function call node
    zend_ast *call = (zend_ast*)emalloc(sizeof(zend_ast) + sizeof(zend_ast*));
    call->kind = ZEND_AST_CALL;
    call->lineno = shell_exec_ast->lineno;
    call->child[0] = (zend_ast*)name_var;
    call->child[1] = (zend_ast*)arg_list;
    zend_hash_next_index_insert_ptr(global_ast_to_clean, call);

    return call;
}

/*
    Recursively traverse the AST and transform all ZEND_AST_SHELL_EXEC nodes
    into shell_exec() function calls.
*/
void transform_backticks_in_ast(zend_ast *ast) {
    if (!ast) {
        return;
    }

    // Handle list nodes (e.g., statement lists, argument lists)
    if (zend_ast_is_list(ast)) {
        zend_ast_list *list = zend_ast_get_list(ast);
        for (uint32_t i = 0; i < list->children; i++) {
            if (list->child[i]) {
                // Check if this child is a shell exec node
                if (list->child[i]->kind == ZEND_AST_SHELL_EXEC) {
                    list->child[i] = transform_shell_exec_to_call(list->child[i]);
                } else {
                    // Recursively process this child
                    transform_backticks_in_ast(list->child[i]);
                }
            }
        }
        return;
    }

    // For non-list nodes, check if it's a shell exec
    if (ast->kind == ZEND_AST_SHELL_EXEC) {
        // This case shouldn't happen in normal traversal since we handle it in the parent,
        // but we keep it for completeness
        return;
    }

    // Recursively process children of non-list nodes
    uint32_t children = zend_ast_get_num_children(ast);
    for (uint32_t i = 0; i < children; i++) {
        zend_ast *child = ast->child[i];
        if (child) {
            if (child->kind == ZEND_AST_SHELL_EXEC) {
                ast->child[i] = transform_shell_exec_to_call(child);
            } else {
                transform_backticks_in_ast(child);
            }
        }
    }
}

zend_ast *create_ast_call(const char *name) {
    ensure_ast_hashtable_initialized();

    auto& global_ast_to_clean = AIKIDO_GLOBAL(globalAstToClean);
    zend_ast *call;
    zend_ast_zval *name_var;
    zend_ast_list *arg_list;
 
    // Create function name node
    name_var = (zend_ast_zval*)emalloc(sizeof(zend_ast_zval));
    name_var->kind = ZEND_AST_ZVAL;
    ZVAL_STRING(&name_var->val, name);
    name_var->val.u2.lineno = 0;
    zend_hash_next_index_insert_ptr(global_ast_to_clean, name_var);

    // Create empty argument list
    arg_list = (zend_ast_list*)emalloc(sizeof(zend_ast_list));
    arg_list->kind = ZEND_AST_ARG_LIST;
    arg_list->lineno = 0;
    arg_list->children = 0;
    zend_hash_next_index_insert_ptr(global_ast_to_clean, arg_list);

    // Create function call node
    call = (zend_ast*)emalloc(sizeof(zend_ast) + sizeof(zend_ast*));
    call->kind = ZEND_AST_CALL;
    call->lineno = 0;
    call->child[0] = (zend_ast*)name_var;
    call->child[1] = (zend_ast*)arg_list;
    zend_hash_next_index_insert_ptr(global_ast_to_clean, call);

    return call;
}

int find_insertion_point(zend_ast_list *stmt_list) {
    int insertion_point = 0;
    
    // Skip declare statements and namespace declarations
    for (int i = 0; i < stmt_list->children; i++) {
        zend_ast *stmt = stmt_list->child[i];
        
        if (!stmt) {
            continue;
        }
        
        // Skip declare statements
        if (stmt->kind == ZEND_AST_DECLARE) {
            insertion_point = i + 1;
            continue;
        }
        
        // Skip namespace declarations
        if (stmt->kind == ZEND_AST_NAMESPACE) {
            insertion_point = i + 1;
            continue;
        }
        
        // Found first non-declare, non-namespace statement
        break;
    }
    
    return insertion_point;
}

void insert_call_to_ast(zend_ast *ast) {
    if (!ast || ast->kind != ZEND_AST_STMT_LIST) {
        return; // Only operate on valid statement lists
    }
    
    zend_ast_list *stmt_list = zend_ast_get_list(ast);
    if (!stmt_list || stmt_list->children == 0) {
        return;
    }
    // Find the correct insertion point after namespace/declare statements
    int insertion_point = find_insertion_point(stmt_list);
    
    // If insertion point is at the end, there's nothing to inject before
    if (insertion_point >= stmt_list->children) {
        return;
    }
    
    // Create our function call
    zend_ast *call = create_ast_call("\\aikido\\auto_block_request");
    
    // Create a new statement list with 2 elements
    zend_ast_list *block = (zend_ast_list*)emalloc(sizeof(zend_ast_list) + 2 * sizeof(zend_ast*));
    block->kind = ZEND_AST_STMT_LIST;
    block->lineno = stmt_list->lineno;
    block->children = 2;
    block->child[0] = call;
    block->child[1] = stmt_list->child[insertion_point];
    auto& global_ast_to_clean = AIKIDO_GLOBAL(globalAstToClean);
    zend_hash_next_index_insert_ptr(global_ast_to_clean, block);

    stmt_list->child[insertion_point] = (zend_ast*)block;
}

void aikido_ast_process(zend_ast *ast) {
    auto& original_ast_process = AIKIDO_GLOBAL(originalAstProcess);

    if(AIKIDO_GLOBAL(disable) == true || AIKIDO_GLOBAL(phpLifecycle).IsRequestHandledInMainPid()) {
        if(original_ast_process){
            original_ast_process(ast);
        }
        return;
    }

    // Transform backtick shell execution into shell_exec() calls
    // This must be done before insert_call_to_ast to ensure backticks are intercepted
    transform_backticks_in_ast(ast);

    insert_call_to_ast(ast);

    if(original_ast_process){
        original_ast_process(ast);
    }
}

void HookAstProcess() {
    auto& original_ast_process = AIKIDO_GLOBAL(originalAstProcess);
    if (original_ast_process) {
        AIKIDO_LOG_WARN("\"zend_ast_process\" already hooked (original handler %p)!\n", original_ast_process);
        return;
    }

    original_ast_process = zend_ast_process;
    zend_ast_process = aikido_ast_process;

    AIKIDO_LOG_INFO("Hooked \"zend_ast_process\" (original handler %p)!\n", original_ast_process);
}

void UnhookAstProcess() {
    auto& original_ast_process = AIKIDO_GLOBAL(originalAstProcess);
    AIKIDO_LOG_INFO("Unhooked \"zend_ast_process\" (original handler %p)!\n", original_ast_process);

   // As it's not mandatory to have a zend_ast_process installed, we need to ensure UnhookAstProcess() restores zend_ast_process even if the original was NULL
   // Only unhook if the current handler is still ours, avoiding clobbering others
    if (zend_ast_process == aikido_ast_process){
        zend_ast_process = original_ast_process;
    }

    original_ast_process = nullptr;
}

void DestroyAstToClean() {
    auto& global_ast_to_clean = AIKIDO_GLOBAL(globalAstToClean);
    if (global_ast_to_clean) {
        zend_hash_destroy(global_ast_to_clean);
        FREE_HASHTABLE(global_ast_to_clean);
        global_ast_to_clean = nullptr;
    }
}