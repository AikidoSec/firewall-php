#include "Includes.h"

HashTable *global_ast_to_clean;
ZEND_API void (*original_ast_process)(zend_ast *ast) = nullptr;

zend_ast *create_ast_call(const char *name)
{
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

void insert_call_to_ast(zend_ast *ast) {
    if (!ast || ast->kind != ZEND_AST_STMT_LIST) {
        return; // Only operate on valid statement lists
    }
    
    zend_ast_list *stmt_list = zend_ast_get_list(ast);
    if (!stmt_list || stmt_list->children == 0) {
        return;
    }
    
    // Create our function call
    zend_ast *call = create_ast_call("aikido\\auto_block_request");
    
    // Create a new statement list with 2 elements
    zend_ast_list *block = (zend_ast_list*)emalloc(sizeof(zend_ast_list) + 2 * sizeof(zend_ast*));
    block->kind = ZEND_AST_STMT_LIST;
    block->lineno = stmt_list->lineno;
    block->children = 2;
    
    // First statement is our call
    block->child[0] = call;
    // Second statement is the original
    block->child[1] = stmt_list->child[0];
    
    // Track the new block for cleanup
    zend_hash_next_index_insert_ptr(global_ast_to_clean, block);
    
    // Replace the first statement with our block
    if (stmt_list->children > 0) {
        stmt_list->child[0] = (zend_ast*)block;
    }
}

void aikido_ast_process(zend_ast *ast)
{
    insert_call_to_ast(ast);

    if(original_ast_process){
        original_ast_process(ast);
    }
}

void ast_to_clean_dtor(zval *zv)
{
    zend_ast *ast = (zend_ast *)Z_PTR_P(zv);
    efree(ast);
} 

void HookZendAstProcess() {
    if (original_ast_process) {
        AIKIDO_LOG_WARN("\"zend_ast_process\" already hooked (original handler %p)!\n", original_ast_process);
        return;
    }

    original_ast_process = zend_ast_process;
    zend_ast_process = aikido_ast_process;

    AIKIDO_LOG_INFO("Hooked \"zend_ast_process\" (original handler %p)!\n", original_ast_process);
}

void UnhookZendAstProcess() {
    if (!original_ast_process) {
        AIKIDO_LOG_WARN("Cannot unhook \"zend_ast_process\" without an original handler (was not previously hooked)!\n");
        return;
    }

    AIKIDO_LOG_INFO("Unhooked \"zend_ast_process\" (original handler %p)!\n", original_ast_process);

    zend_ast_process = original_ast_process;
    original_ast_process = nullptr;
}

void InitAstToClean() {
    ALLOC_HASHTABLE(global_ast_to_clean);
    zend_hash_init(global_ast_to_clean, 8, NULL, ast_to_clean_dtor, 1);
}

void DestroyAstToClean() {
    zend_hash_destroy(global_ast_to_clean);
    FREE_HASHTABLE(global_ast_to_clean);
}