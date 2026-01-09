#include "Includes.h"

ZEND_FUNCTION(register_param_matcher) {
    ScopedTimer scopedTimer("register_param_matcher", "aikido_op");
    ScopedEventContext scopedContext; // Push context for eventCacheStack

    if (IsAikidoDisabledOrBypassed()) {
        RETURN_BOOL(false);
    }

    char *param = nullptr;
    size_t paramLength = 0;
    char *regex = nullptr;
    size_t regexLength = 0;

    ZEND_PARSE_PARAMETERS_START(2, 2)
        Z_PARAM_STRING(param, paramLength)
        Z_PARAM_STRING(regex, regexLength)
    ZEND_PARSE_PARAMETERS_END();

    if (!param || paramLength == 0 || !regex || regexLength == 0) {
        AIKIDO_LOG_ERROR("register_param_matcher: param or regex is null or empty!\n");
        RETURN_BOOL(false);
    }

    eventCacheStack.Current().paramMatcherParam = std::string(param, paramLength);
    eventCacheStack.Current().paramMatcherRegex = std::string(regex, regexLength);

    try {
        std::string outputEvent;
        requestProcessor.SendEvent(EVENT_REGISTER_PARAM_MATCHER, outputEvent);
        if (action.Execute(outputEvent) == WARNING_MESSAGE) {
            // If a warning message is returned, it means that the parameters are invalid, so we return false.
            RETURN_BOOL(false);
        }

    } catch (const std::exception& e) {
        AIKIDO_LOG_ERROR("Exception encountered in processing register param matcher event: %s\n", e.what());
        RETURN_BOOL(false);
    }

    AIKIDO_LOG_INFO("Registered param matcher %s -> %s\n", eventCacheStack.Current().paramMatcherParam.c_str(), eventCacheStack.Current().paramMatcherRegex.c_str());
    RETURN_BOOL(true);
}
