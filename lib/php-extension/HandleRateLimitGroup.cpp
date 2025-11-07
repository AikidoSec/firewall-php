#include "Includes.h"

ZEND_FUNCTION(set_rate_limit_group) {
    if (AIKIDO_GLOBAL(disable) == true) {
        RETURN_BOOL(false);
    }

    char* group = nullptr;
    size_t groupLength = 0;

    ZEND_PARSE_PARAMETERS_START(1, 1)
        Z_PARAM_STRING(group, groupLength)
    ZEND_PARSE_PARAMETERS_END();

    if (group == nullptr || groupLength == 0) {
        AIKIDO_LOG_ERROR("set_rate_limit_group: group is null!\n");
        RETURN_BOOL(false);
    }

    auto& requestCache = AIKIDO_GLOBAL(requestCache);
    requestCache.rateLimitGroup = std::string(group, groupLength);

    std::string outputEvent;
    AIKIDO_GLOBAL(requestProcessor).SendEvent(EVENT_SET_RATE_LIMIT_GROUP, outputEvent);
    AIKIDO_LOG_DEBUG("Set rate limit group to %s\n", requestCache.rateLimitGroup.c_str());

    RETURN_BOOL(true);
}