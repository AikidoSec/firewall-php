#include "Includes.h"

void PhpLifecycle::ModuleInit() {
    this->mainPID = getpid();
    AIKIDO_LOG_INFO("Main PID is: %u\n", this->mainPID);
    if (!AIKIDO_GLOBAL(agent).Init()) {
        AIKIDO_LOG_INFO("Aikido Agent initialization failed!\n");
    } else {
        AIKIDO_LOG_INFO("Aikido Agent initialization succeeded!\n");
    }
}

void PhpLifecycle::RequestInit() {
    action.Reset();
    requestCache.Reset();
    requestProcessor.RequestInit();
    checkedAutoBlock = false;
    checkedShouldBlockRequest = false;
}

void PhpLifecycle::RequestShutdown() {
    requestProcessor.RequestShutdown();
}

void PhpLifecycle::ModuleShutdown() {
    char* v = sapi_getenv("SERVER_SOFTWARE", sizeof("SERVER_SOFTWARE")-1)
    if (this->mainPID == getpid()) {
        if (AIKIDO_GLOBAL(sapi_name) == "fpm-fcgi") {
            AIKIDO_LOG_INFO("Module shutdown called on main PID for php-fpm (server software: %s). Ignoring...\n", v);
            return;
        }

        AIKIDO_LOG_INFO("Module shutdown called on main PID.\n");
        AIKIDO_LOG_INFO("Unhooking functions...\n");
        AIKIDO_LOG_INFO("Uninitializing Aikido Agent...\n");
        AIKIDO_GLOBAL(agent).Uninit();
        UnhookAll();
    } else {
        AIKIDO_LOG_INFO("Module shutdown NOT called on main PID (server software: %s). Uninitializing Aikido Request Processor...\n", v);
        requestProcessor.Uninit();
    }
}

void PhpLifecycle::HookAll() {
    HookFunctions();
    HookMethods();
    HookFileCompilation();
    HookAstProcess();
}

void PhpLifecycle::UnhookAll() {
    UnhookFunctions();
    UnhookMethods();
    UnhookFileCompilation();
    UnhookAstProcess();
}

PhpLifecycle phpLifecycle;
