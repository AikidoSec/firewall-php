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
    AIKIDO_GLOBAL(action).Reset();
    AIKIDO_GLOBAL(requestCache).Reset();
    AIKIDO_GLOBAL(requestProcessor).RequestInit();
    AIKIDO_GLOBAL(checkedAutoBlock) = false;
    AIKIDO_GLOBAL(checkedShouldBlockRequest) = false;
    AIKIDO_GLOBAL(checkedIpBypass) = false;
    AIKIDO_GLOBAL(isIpBypassed) = false;
}

void PhpLifecycle::RequestShutdown() {
    AIKIDO_GLOBAL(requestProcessor).RequestShutdown();
}

void PhpLifecycle::ModuleShutdown() {
#ifdef ZTS
    AIKIDO_LOG_INFO("ZTS mode: Uninitializing Aikido Request Processor to stop background goroutines...\n");
    AIKIDO_GLOBAL(requestProcessor).Uninit();
#else
    if (this->mainPID == getpid()) {
        AIKIDO_LOG_INFO("Module shutdown called on main PID.\n");
        AIKIDO_LOG_INFO("Unhooking functions...\n");
        AIKIDO_LOG_INFO("Uninitializing Aikido Agent...\n");
        AIKIDO_GLOBAL(agent).Uninit();
        UnhookAll();
    } else {
        AIKIDO_LOG_INFO("Module shutdown NOT called on main PID. Uninitializing Aikido Request Processor...\n");
        AIKIDO_GLOBAL(requestProcessor).Uninit();
    }
#endif
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
