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

    AIKIDO_GLOBAL(requestProcessorInstance).RequestInit();
    AIKIDO_GLOBAL(checkedAutoBlock) = false;
    AIKIDO_GLOBAL(checkedShouldBlockRequest) = false;
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
        AIKIDO_LOG_INFO("Module shutdown NOT called on main PID. Uninitializing Aikido Request Processor...\n");
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
