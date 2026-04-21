#include "Includes.h"

void PhpLifecycle::ModuleInit() {
    /* If SAPI name is "cli" run in "simple" mode */
    if (AIKIDO_GLOBAL(sapi_name) == "cli") {
        AIKIDO_LOG_INFO("MINIT finished earlier because we run in CLI mode!\n");
        return;
    }

    this->mainPID = getpid();
    AIKIDO_LOG_INFO("Main PID is: %u\n", this->mainPID);
    if (!AIKIDO_GLOBAL(agent).Init()) {
        AIKIDO_LOG_INFO("Aikido Agent initialization failed!\n");
    } else {
        AIKIDO_LOG_INFO("Aikido Agent initialization succeeded!\n");
    }
}

void PhpLifecycle::RequestInit() {
    // Skip RequestInit in the php-fpm master's opcache.preload virtual RINIT
    // cycle: running it there would dlopen aikido-request-processor.so in
    // the master, and every forked worker would inherit a Go runtime whose
    // scheduler threads do not exist in its address space.
   #ifdef NTS
    if (this->mainPID == getpid() && AIKIDO_GLOBAL(sapi_name) == "fpm-fcgi"){
        AIKIDO_LOG_INFO("Skipping RequestInit in php-fpm master (pid %d == mainPID; "
                        "likely opcache.preload virtual RINIT). Workers will initialize "
                        "the request processor after fork.\n", (int)getpid());
        return;
    }
    #endif

    AIKIDO_GLOBAL(action).Reset();
    AIKIDO_GLOBAL(requestCache).Reset();
    
    AIKIDO_GLOBAL(requestProcessorInstance).RequestInit();
    AIKIDO_GLOBAL(checkedAutoBlock) = false;
    AIKIDO_GLOBAL(checkedShouldBlockRequest) = false;
    AIKIDO_GLOBAL(checkedWhitelistRequest) = false;
    AIKIDO_GLOBAL(isIpBypassed) = false;
    InitIpBypassCheck();
}

void PhpLifecycle::RequestShutdown() {
    AIKIDO_GLOBAL(requestProcessorInstance).RequestShutdown();
}

void PhpLifecycle::ModuleShutdown() {
    if (this->mainPID == getpid()) {
        AIKIDO_LOG_INFO("Module shutdown called on main PID.\n");
        AIKIDO_LOG_INFO("Unhooking functions...\n");
        AIKIDO_LOG_INFO("Uninitializing Aikido Agent...\n");
        AIKIDO_GLOBAL(agent).Uninit();
        UnhookAll();
    } else {
        #ifndef ZTS
            AIKIDO_LOG_INFO("Module shutdown NOT called on main PID. Uninitializing Aikido Request Processor...\n");
            requestProcessor.Uninit();
        #endif
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
