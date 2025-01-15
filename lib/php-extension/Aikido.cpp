/* Aikido runtime extension for PHP */
#include "Includes.h"

ZEND_DECLARE_MODULE_GLOBALS(aikido)

PHP_MINIT_FUNCTION(aikido) {
    LoadEnvironment();
    AIKIDO_GLOBAL(socket_path) = GenerateSocketPath();
    AIKIDO_GLOBAL(logger).Init();

    AIKIDO_LOG_INFO("MINIT started!\n");

    RegisterAikidoBlockRequestStatusClass();

    if (AIKIDO_GLOBAL(disable) == true) {
        AIKIDO_LOG_INFO("MINIT finished earlier because AIKIDO_DISABLE is set to 1!\n");
        return SUCCESS;
    }

    HookFunctions();
    HookMethods();
    HookFileCompilation();

    /* If SAPI name is "cli" run in "simple" mode */
    if (AIKIDO_GLOBAL(sapi_name) == "cli") {
        AIKIDO_LOG_INFO("MINIT finished earlier because we run in CLI mode!\n");
        return SUCCESS;
    }

    phpLifecycle.ModuleInit();
    AIKIDO_LOG_INFO("MINIT finished!\n");
    return SUCCESS;
}

PHP_MSHUTDOWN_FUNCTION(aikido) {
    AIKIDO_LOG_DEBUG("MSHUTDOWN started!\n");

    LogScopedUninit logScopedUninit;

    if (AIKIDO_GLOBAL(disable) == true) {
        AIKIDO_LOG_INFO("MSHUTDOWN finished earlier because AIKIDO_DISABLE is set to 1!\n");
        return SUCCESS;
    }

    /* If SAPI name is "cli" run in "simple" mode */
    if (AIKIDO_GLOBAL(sapi_name) == "cli") {
        AIKIDO_LOG_INFO("MSHUTDOWN finished earlier because we run in CLI mode!\n");
        return SUCCESS;
    }

    phpLifecycle.ModuleShutdown();
    AIKIDO_LOG_DEBUG("MSHUTDOWN finished!\n");
    return SUCCESS;
}

PHP_RINIT_FUNCTION(aikido) {
    ScopedTimer scopedTimer("request_init");

    AIKIDO_LOG_DEBUG("RINIT started!\n");
    
    if (AIKIDO_GLOBAL(disable) == true) {
        AIKIDO_LOG_INFO("RINIT finished earlier because AIKIDO_DISABLE is set to 1!\n");
        return SUCCESS;
    }

    phpLifecycle.RequestInit();
    AIKIDO_LOG_DEBUG("RINIT finished!\n");
    return SUCCESS;
}

PHP_RSHUTDOWN_FUNCTION(aikido) {
    ScopedTimer scopedTimer("request_shutdown");

    AIKIDO_LOG_DEBUG("RSHUTDOWN started!\n");

    if (AIKIDO_GLOBAL(disable) == true) {
        AIKIDO_LOG_INFO("RSHUTDOWN finished earlier because AIKIDO_DISABLE is set to 1!\n");
        return SUCCESS;
    }

    phpLifecycle.RequestShutdown();
    AIKIDO_LOG_DEBUG("RSHUTDOWN finished!\n");
    return SUCCESS;
}

PHP_MINFO_FUNCTION(aikido) {
    php_info_print_table_start();
    php_info_print_table_row(2, "aikido support", "enabled");
    php_info_print_table_end();
}

static const zend_function_entry ext_functions[] = {
    ZEND_NS_FE("aikido", set_user, arginfo_aikido_set_user)
        ZEND_NS_FE("aikido", should_block_request, arginfo_aikido_should_block_request)
            ZEND_FE_END};

zend_module_entry aikido_module_entry = {
    STANDARD_MODULE_HEADER,
    "aikido",                   /* Extension name */
    ext_functions,              /* zend_function_entry */
    PHP_MINIT(aikido),          /* PHP_MINIT - Module initialization */
    PHP_MSHUTDOWN(aikido),      /* PHP_MSHUTDOWN - Module shutdown */
    PHP_RINIT(aikido),          /* PHP_RINIT - Request initialization */
    PHP_RSHUTDOWN(aikido),      /* PHP_RSHUTDOWN - Request shutdown */
    PHP_MINFO(aikido),          /* PHP_MINFO - Module info */
    PHP_AIKIDO_VERSION,         /* Version */
    PHP_MODULE_GLOBALS(aikido), /* Module globals */
    NULL,                       /* PHP_GINIT – Globals initialization */
    NULL,                       /* PHP_GSHUTDOWN – Globals shutdown */
    NULL,
    STANDARD_MODULE_PROPERTIES_EX,
};

#ifdef COMPILE_DL_AIKIDO
#ifdef ZTS
ZEND_TSRMLS_CACHE_DEFINE()
#endif
ZEND_GET_MODULE(aikido)
#endif
