/* Aikido runtime extension for PHP */
#include "Includes.h"

ZEND_DECLARE_MODULE_GLOBALS(aikido)

PHP_MINIT_FUNCTION(aikido) {
    // For FrankenPHP: Set sapi_name but skip rest of LoadEnvironment during MINIT
    // Full environment will be loaded in RINIT when Caddyfile env vars are available
    if (sapi_module.name == std::string("frankenphp")) {
        AIKIDO_GLOBAL(sapi_name) = sapi_module.name;
    } else {
        // For other SAPIs: Load environment during MINIT as normal
        LoadEnvironment();
    }
    AIKIDO_GLOBAL(logger).Init();

    AIKIDO_LOG_INFO("MINIT started!\n");

    RegisterAikidoBlockRequestStatusClass();

    if (AIKIDO_GLOBAL(disable) == true) {
        AIKIDO_LOG_INFO("MINIT finished earlier because AIKIDO_DISABLE is set to 1!\n");
        return SUCCESS;
    }

    AIKIDO_GLOBAL(phpLifecycle).HookAll();
    /* If SAPI name is "cli" run in "simple" mode */
    if (AIKIDO_GLOBAL(sapi_name) == "cli") {
        AIKIDO_LOG_INFO("MINIT finished earlier because we run in CLI mode!\n");
        return SUCCESS;
    }

    AIKIDO_GLOBAL(phpLifecycle).ModuleInit();
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
        AIKIDO_GLOBAL(phpLifecycle).UnhookAll();
        return SUCCESS;
    }

    /*
        In the case of Apache mod-php servers, the MSHUTDOWN can be called multiple times.
        As a consequence, we need to do the unhooking / uninitialization logic based on the current
        PID for which the MSHUTDOWN is called. This logic is part of phpLifecycle.ModuleShutdown().
        The same does not apply for CLI mode, where the MSHUTDOWN is called only once.
    */

    AIKIDO_GLOBAL(phpLifecycle).ModuleShutdown();
    AIKIDO_LOG_DEBUG("MSHUTDOWN finished!\n");
    return SUCCESS;
}
// Common request initialization logic shared between RINIT and aikido_worker_rinit
static void aikido_do_request_init() {
    AIKIDO_GLOBAL(phpLifecycle).RequestInit();
}

// Common request shutdown logic shared between RSHUTDOWN and aikidoworker_rshutdown
static void aikido_do_request_shutdown() {
    if (AIKIDO_GLOBAL(disable) != true) {
        DestroyAstToClean();
        AIKIDO_GLOBAL(phpLifecycle).RequestShutdown();
    }
}

PHP_RINIT_FUNCTION(aikido) {
    ScopedTimer scopedTimer("request_init", "request_op");
    
    aikido_do_request_init();

    AIKIDO_LOG_DEBUG("RINIT finished!\n");
    return SUCCESS;
}

PHP_RSHUTDOWN_FUNCTION(aikido) {
    ScopedTimer scopedTimer("request_shutdown", "request_op");

    AIKIDO_LOG_DEBUG("RSHUTDOWN started!\n");

    aikido_do_request_shutdown();
    
    AIKIDO_LOG_DEBUG("RSHUTDOWN finished!\n");
    return SUCCESS;
}

// PHP function: \aikido\worker_rinit()
// Because FrankenPHP doesn't call RINIT for each request in worker mode, 
// we need to call it manually at the start of each request.
// Only works with FrankenPHP worker mode
PHP_FUNCTION(worker_rinit) {
    ZEND_PARSE_PARAMETERS_NONE();
    
    // Only allow this function in FrankenPHP worker mode
    if (std::string(sapi_module.name) != "frankenphp" || !AIKIDO_GLOBAL(isWorkerMode)) {
        zend_throw_exception(
            GetFirewallDefaultExceptionCe(),
            "aikido\\worker_rinit() can only be called in FrankenPHP worker mode", 0);
        RETURN_FALSE;
    }
    
    aikido_do_request_init();
    
    RETURN_TRUE;
}

// PHP function: \aikido\worker_rshutdown()
// Because FrankenPHP doesn't call RSHUTDOWN for each request in worker mode,
// we need to call it manually at the end of each request.
// Only works with FrankenPHP worker mode
PHP_FUNCTION(worker_rshutdown) {
    ZEND_PARSE_PARAMETERS_NONE();
    
    // Only allow this function in FrankenPHP worker mode
    if (std::string(sapi_module.name) != "frankenphp" || !AIKIDO_GLOBAL(isWorkerMode)) {
        zend_throw_exception(
            GetFirewallDefaultExceptionCe(),
            "aikido\\worker_rshutdown() can only be called in FrankenPHP worker mode", 0);
        RETURN_FALSE;
    }
    
    aikido_do_request_shutdown();
    
    RETURN_TRUE;
}

PHP_MINFO_FUNCTION(aikido) {
    php_info_print_table_start();
    php_info_print_table_row(2, "aikido support", "enabled");
    php_info_print_table_end();
}

static const zend_function_entry ext_functions[] = {
    ZEND_NS_FE("aikido", set_user, arginfo_aikido_set_user)
    ZEND_NS_FE("aikido", should_block_request, arginfo_aikido_should_block_request)
    ZEND_NS_FE("aikido", auto_block_request, arginfo_aikido_auto_block_request)
    ZEND_NS_FE("aikido", set_token, arginfo_aikido_set_token)
    ZEND_NS_FE("aikido", set_rate_limit_group, arginfo_aikido_set_rate_limit_group)
    ZEND_NS_FE("aikido", register_param_matcher, arginfo_aikido_register_param_matcher)
    ZEND_NS_FE("aikido", worker_rinit, arginfo_aikido_worker_rinit)
    ZEND_NS_FE("aikido", worker_rshutdown, arginfo_aikido_worker_rshutdown)
    ZEND_FE_END
};

PHP_GINIT_FUNCTION(aikido) {
    aikido_globals->environment_loaded = false;
    aikido_globals->log_level = 0;
    aikido_globals->blocking = false;
    aikido_globals->disable = false;
    aikido_globals->disk_logs = false;
    aikido_globals->collect_api_schema = false;
    aikido_globals->trust_proxy = false;
    aikido_globals->localhost_allowed_by_default = false;
    aikido_globals->report_stats_interval_to_agent = 0;
    aikido_globals->currentRequestStart = std::chrono::high_resolution_clock::time_point{};
    aikido_globals->totalOverheadForCurrentRequest = 0;
    aikido_globals->laravelEnvLoaded = false;
    aikido_globals->checkedAutoBlock = false;
    aikido_globals->checkedShouldBlockRequest = false;
    aikido_globals->checkedIpBypass = false;
    aikido_globals->isIpBypassed = false;
    aikido_globals->isWorkerMode = false;
    aikido_globals->global_ast_to_clean = nullptr;
    aikido_globals->original_ast_process = nullptr;
#ifdef ZTS
    new (&aikido_globals->log_level_str) std::string();
    new (&aikido_globals->sapi_name) std::string();
    new (&aikido_globals->token) std::string();
    new (&aikido_globals->endpoint) std::string();
    new (&aikido_globals->config_endpoint) std::string();
    new (&aikido_globals->logger) Log();
    new (&aikido_globals->agent) Agent();
    new (&aikido_globals->server) Server();
    new (&aikido_globals->requestProcessor) RequestProcessor();
    new (&aikido_globals->action) Action();
    new (&aikido_globals->requestCache) RequestCache();
    new (&aikido_globals->eventCacheStack) EventCacheStack();
    new (&aikido_globals->eventCache) EventCache();
    new (&aikido_globals->phpLifecycle) PhpLifecycle();
    new (&aikido_globals->stats) std::unordered_map<std::string, SinkStats>();
    new (&aikido_globals->laravelEnv) std::unordered_map<std::string, std::string>();
#endif
}

PHP_GSHUTDOWN_FUNCTION(aikido) {
#ifdef ZTS
    aikido_globals->laravelEnv.~unordered_map();
    aikido_globals->phpLifecycle.~PhpLifecycle();
    aikido_globals->action.~Action();
    aikido_globals->requestProcessor.~RequestProcessor();
    aikido_globals->stats.~unordered_map();
    aikido_globals->server.~Server();
    aikido_globals->logger.~Log();
    aikido_globals->agent.~Agent();
    aikido_globals->eventCache.~EventCache();
    aikido_globals->eventCacheStack.~EventCacheStack();
    aikido_globals->requestCache.~RequestCache();
    aikido_globals->config_endpoint.~string();
    aikido_globals->endpoint.~string();
    aikido_globals->token.~string();
    aikido_globals->sapi_name.~string();
    aikido_globals->log_level_str.~string();
#endif
}

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
    PHP_GINIT(aikido),          /* PHP_GINIT – Globals initialization */
    PHP_GSHUTDOWN(aikido),      /* PHP_GSHUTDOWN – Globals shutdown */
    NULL,
    STANDARD_MODULE_PROPERTIES_EX,
};

#ifdef COMPILE_DL_AIKIDO
#ifdef ZTS
ZEND_TSRMLS_CACHE_DEFINE()
#endif
ZEND_GET_MODULE(aikido)
#endif
