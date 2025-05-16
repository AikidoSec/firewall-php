#pragma once


extern HashTable *global_ast_to_clean;
extern ZEND_API void (*original_ast_process)(zend_ast *ast);

extern bool checkedAutoBlock;
extern bool checkedShouldBlockRequest;

void HookAstProcess();
void UnhookAstProcess();
void DestroyAstToClean();
