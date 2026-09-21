#include "Includes.h"

ZEND_FUNCTION(track) {
    if (AIKIDO_GLOBAL(sapi_name) == "cli" || IsAikidoDisabledOrBypassed()) {
        RETURN_BOOL(false);
    }

    char* eventName = nullptr;
    size_t eventNameLength = 0;

    ZEND_PARSE_PARAMETERS_START(1, 1)
        Z_PARAM_STRING(eventName, eventNameLength)
    ZEND_PARSE_PARAMETERS_END();

    if (!eventName || eventNameLength == 0) {
        AIKIDO_LOG_INFO("track(...) expects a non-empty string as event name.\n");
        RETURN_BOOL(false);
    }

    ScopedTimer scopedTimer("track", "aikido_op");

    try {
        ScopedEventContext scopedContext;
        AIKIDO_GLOBAL(eventCacheStack).Top().customEventName = std::string(eventName, eventNameLength);

        std::string outputEvent;
        RETURN_BOOL(AIKIDO_GLOBAL(requestProcessorInstance).SendEvent(EVENT_TRACK, outputEvent));
    } catch (const std::exception& e) {
        AIKIDO_LOG_DEBUG("Exception encountered while tracking custom event: %s\n", e.what());
        RETURN_BOOL(false);
    }
}
