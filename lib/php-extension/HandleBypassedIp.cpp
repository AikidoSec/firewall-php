#include "Includes.h"

void InitIpBypassCheck() {
    // Reset state for new request
    AIKIDO_GLOBAL(isIpBypassed) = false;

    ScopedTimer scopedTimer("check_ip_bypass", "aikido_op");

    try {
        std::string output;
        AIKIDO_GLOBAL(requestProcessorInstance).SendEvent(EVENT_GET_IS_IP_BYPASSED, output);
        AIKIDO_GLOBAL(action).Execute(output);
    } catch (const std::exception &e) {
        AIKIDO_LOG_ERROR("Exception encountered in processing IP bypass check event: %s\n", e.what());
    }
}


bool IsAikidoDisabledOrBypassed() {
    return AIKIDO_GLOBAL(disable) == true || AIKIDO_GLOBAL(isIpBypassed);
}

