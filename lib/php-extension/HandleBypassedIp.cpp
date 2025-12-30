#include "Includes.h"

// The isIpBypassed module global variable is used to store whether the current IP is bypassed.
// If true, all blocking checks will be skipped.
// Accessed via AIKIDO_GLOBAL(isIpBypassed).

// The checkedIpBypass module global variable is used to check if IP bypass check
// has already been called, in order to avoid multiple calls to this function.
// Accessed via AIKIDO_GLOBAL(checkedIpBypass).

void InitIpBypassCheck() {
    ScopedTimer scopedTimer("check_ip_bypass", "aikido_op");

    try {
        std::string output;
        AIKIDO_GLOBAL(requestProcessor).SendEvent(EVENT_GET_IS_IP_BYPASSED, output);
        AIKIDO_GLOBAL(action).Execute(output);
    } catch (const std::exception &e) {
        AIKIDO_LOG_ERROR("Exception encountered in processing IP bypass check event: %s\n", e.what());
    }
}

bool IsAikidoDisabledOrBypassed() {
    if (AIKIDO_GLOBAL(disable) == true) {
        return true;
    }
    
    if (!AIKIDO_GLOBAL(checkedIpBypass)) {
        AIKIDO_GLOBAL(checkedIpBypass) = true;
        InitIpBypassCheck();
    }
    
    return AIKIDO_GLOBAL(isIpBypassed);
}

