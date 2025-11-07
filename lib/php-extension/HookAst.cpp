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
    auto& globalAstToClean = AIKIDO_GLOBAL(global_ast_to_clean);
    if (!globalAstToClean) {
        ALLOC_HASHTABLE(globalAstToClean);
        zend_hash_init(globalAstToClean, 8, NULL, ast_to_clean_dtor, 1);
    }
}

zend_ast *create_ast_call(const char *name) {
    ensure_ast_hashtable_initialized();

    auto& globalAstToClean = AIKIDO_GLOBAL(global_ast_to_clean);
    zend_ast *call;
    zend_ast_zval *name_var;
    zend_ast_list *arg_list;
 
    // Create function name node
    name_var = (zend_ast_zval*)emalloc(sizeof(zend_ast_zval));
    name_var->kind = ZEND_AST_ZVAL;
    ZVAL_STRING(&name_var->val, name);
    name_var->val.u2.lineno = 0;
    zend_hash_next_index_insert_ptr(globalAstToClean, name_var);

    // Create empty argument list
    arg_list = (zend_ast_list*)emalloc(sizeof(zend_ast_list));
    arg_list->kind = ZEND_AST_ARG_LIST;
    arg_list->lineno = 0;
    arg_list->children = 0;
    zend_hash_next_index_insert_ptr(globalAstToClean, arg_list);

    // Create function call node
    call = (zend_ast*)emalloc(sizeof(zend_ast) + sizeof(zend_ast*));
    call->kind = ZEND_AST_CALL;
    call->lineno = 0;
    call->child[0] = (zend_ast*)name_var;
    call->child[1] = (zend_ast*)arg_list;
    zend_hash_next_index_insert_ptr(globalAstToClean, call);

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
    auto& globalAstToClean = AIKIDO_GLOBAL(global_ast_to_clean);
    zend_hash_next_index_insert_ptr(globalAstToClean, block);

    stmt_list->child[insertion_point] = (zend_ast*)block;
}

void aikido_ast_process(zend_ast *ast) {
    insert_call_to_ast(ast);

    auto& originalAstProcess = AIKIDO_GLOBAL(original_ast_process);
    if(originalAstProcess){
        originalAstProcess(ast);
    }
}

void HookAstProcess() {
    auto& originalAstProcess = AIKIDO_GLOBAL(original_ast_process);
    if (originalAstProcess) {
        AIKIDO_LOG_WARN("\"zend_ast_process\" already hooked (original handler %p)!\n", originalAstProcess);
        return;
    }

    originalAstProcess = zend_ast_process;
    zend_ast_process = aikido_ast_process;

    AIKIDO_LOG_INFO("Hooked \"zend_ast_process\" (original handler %p)!\n", originalAstProcess);
}

void UnhookAstProcess() {
    auto& originalAstProcess = AIKIDO_GLOBAL(original_ast_process);
    AIKIDO_LOG_INFO("Unhooked \"zend_ast_process\" (original handler %p)!\n", originalAstProcess);

   // As it's not mandatory to have a zend_ast_process installed, we need to ensure UnhookAstProcess() restores zend_ast_process even if the original was NULL
   // Only unhook if the current handler is still ours, avoiding clobbering others
    if (zend_ast_process == aikido_ast_process){
        zend_ast_process = originalAstProcess;
    }

    originalAstProcess = nullptr;
}

void DestroyAstToClean() {
    auto& globalAstToClean = AIKIDO_GLOBAL(global_ast_to_clean);
    if (globalAstToClean) {
        zend_hash_destroy(globalAstToClean);
        FREE_HASHTABLE(globalAstToClean);
        globalAstToClean = nullptr;
    }
}