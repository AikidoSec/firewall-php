#include "Includes.h"

HashTable *global_ast_to_clean;
ZEND_API void (*original_ast_process)(zend_ast *ast) = nullptr;

/*
    This is a custom destructor, used for cleaning the allocated ast hashtable.
    This is needed because the ast hashtable is not cleaned by the zend_ast_process function.
*/
void ast_to_clean_dtor(zval *zv) {
    zend_ast *ast = (zend_ast *)Z_PTR_P(zv);
    efree(ast);
} 

void ensure_ast_hashtable_initialized() {
    if (!global_ast_to_clean) {
        ALLOC_HASHTABLE(global_ast_to_clean);
        zend_hash_init(global_ast_to_clean, 8, NULL, ast_to_clean_dtor, 1);
    }
}

zend_ast *create_ast_call(const char *name) {
    ensure_ast_hashtable_initialized();

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

/*
    This function creates an if statement that checks if the Aikido extension is loaded.
    If it is, it calls the auto_block_request function.
    If it is not, it does nothing.
    This is used to avoid calling the auto_block_request if the extension is not loaded anymore.
    
    if (extension_loaded('aikido')) {
        \aikido\auto_block_request();
    }
*/
zend_ast *create_ast_if_aikido() {
    ensure_ast_hashtable_initialized();
    
    // Create the function call: \aikido\auto_block_request()
    zend_ast *call = create_ast_call("\\aikido\\auto_block_request");
    
    // Create the extension_loaded('aikido') call
    zend_ast_zval *ext_name;
    zend_ast_list *ext_arg_list;
    zend_ast *ext_call;
    
    // Create 'aikido' string argument
    ext_name = (zend_ast_zval*)emalloc(sizeof(zend_ast_zval));
    ext_name->kind = ZEND_AST_ZVAL;
    ZVAL_STRING(&ext_name->val, "aikido");
    ext_name->val.u2.lineno = 0;
    zend_hash_next_index_insert_ptr(global_ast_to_clean, ext_name);
    
    // Create argument list for extension_loaded
    ext_arg_list = (zend_ast_list*)emalloc(sizeof(zend_ast_list) + sizeof(zend_ast*));
    ext_arg_list->kind = ZEND_AST_ARG_LIST;
    ext_arg_list->lineno = 0;
    ext_arg_list->children = 1;
    ext_arg_list->child[0] = (zend_ast*)ext_name;
    zend_hash_next_index_insert_ptr(global_ast_to_clean, ext_arg_list);
    
    // Create extension_loaded function name
    zend_ast_zval *ext_func_name;
    ext_func_name = (zend_ast_zval*)emalloc(sizeof(zend_ast_zval));
    ext_func_name->kind = ZEND_AST_ZVAL;
    ZVAL_STRING(&ext_func_name->val, "extension_loaded");
    ext_func_name->val.u2.lineno = 0;
    zend_hash_next_index_insert_ptr(global_ast_to_clean, ext_func_name);
    
    // Create extension_loaded() call
    ext_call = (zend_ast*)emalloc(sizeof(zend_ast) + sizeof(zend_ast*));
    ext_call->kind = ZEND_AST_CALL;
    ext_call->lineno = 0;
    ext_call->child[0] = (zend_ast*)ext_func_name;
    ext_call->child[1] = (zend_ast*)ext_arg_list;
    zend_hash_next_index_insert_ptr(global_ast_to_clean, ext_call);
    
    // Create statement list for the if body
    zend_ast_list *stmt_list;
    stmt_list = (zend_ast_list*)emalloc(sizeof(zend_ast_list) + sizeof(zend_ast*));
    stmt_list->kind = ZEND_AST_STMT_LIST;
    stmt_list->lineno = 0;
    stmt_list->children = 1;
    stmt_list->child[0] = call;
    zend_hash_next_index_insert_ptr(global_ast_to_clean, stmt_list);
    
    // Create the if statement
    zend_ast_list *if_stmt;
    if_stmt = (zend_ast_list*)emalloc(sizeof(zend_ast_list) + (2 * sizeof(zend_ast*)));
    if_stmt->kind = ZEND_AST_IF;
    if_stmt->lineno = 0;
    if_stmt->children = 1;
    
    // Create if element (condition + body)
    zend_ast *if_elem;
    if_elem = (zend_ast*)emalloc(sizeof(zend_ast) + sizeof(zend_ast*));
    if_elem->kind = ZEND_AST_IF_ELEM;
    if_elem->lineno = 0;
    if_elem->child[0] = ext_call;  // condition
    if_elem->child[1] = (zend_ast*)stmt_list;  // body
    zend_hash_next_index_insert_ptr(global_ast_to_clean, if_elem);
    
    if_stmt->child[0] = if_elem;
    zend_hash_next_index_insert_ptr(global_ast_to_clean, if_stmt);
    
    return (zend_ast*)if_stmt;
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
    zend_ast *call = create_ast_if_aikido();
    // Create a new statement list with 2 elements
    zend_ast_list *block = (zend_ast_list*)emalloc(sizeof(zend_ast_list) + 2 * sizeof(zend_ast*));
    block->kind = ZEND_AST_STMT_LIST;
    block->lineno = stmt_list->lineno;
    block->children = 2;
    block->child[0] = call;
    block->child[1] = stmt_list->child[insertion_point];
    zend_hash_next_index_insert_ptr(global_ast_to_clean, block);

    stmt_list->child[insertion_point] = (zend_ast*)block;
}

void aikido_ast_process(zend_ast *ast) {
    insert_call_to_ast(ast);

    if(original_ast_process){
        original_ast_process(ast);
    }
}

void HookAstProcess() {
    if (original_ast_process) {
        AIKIDO_LOG_WARN("\"zend_ast_process\" already hooked (original handler %p)!\n", original_ast_process);
        return;
    }

    original_ast_process = zend_ast_process;
    zend_ast_process = aikido_ast_process;

    AIKIDO_LOG_INFO("Hooked \"zend_ast_process\" (original handler %p)!\n", original_ast_process);
}

void UnhookAstProcess() {
    AIKIDO_LOG_INFO("Unhooked \"zend_ast_process\" (original handler %p)!\n", original_ast_process);

   // As it's not mandatory to have a zend_ast_process installed, we need to ensure UnhookAstProcess() restores zend_ast_process even if the original was NULL
   // Only unhook if the current handler is still ours, avoiding clobbering others
    if (zend_ast_process == aikido_ast_process){
        zend_ast_process = original_ast_process;
    }

    original_ast_process = nullptr;
}

void DestroyAstToClean() {
    if (global_ast_to_clean) {
        zend_hash_destroy(global_ast_to_clean);
        FREE_HASHTABLE(global_ast_to_clean);
        global_ast_to_clean = nullptr;
    }
}