#include "Includes.h"

/* Helper for handle pre file path access */
void helper_handle_pre_file_path_access(char *filename, EVENT_ID &eventId) {
    if (strncmp(filename, "php://", 6) == 0 && 
        !StartsWithCaseInsensitive(filename, "php://filter")) {
        // Whitelist all php:// streams apart from php://filter, for performance reasons (some PHP frameworks do 1000+ calls / request with these streams as param)
        // php://filter can be used to open arbitrary files, so we still monitor this
        return;
    }

    // if filename starts with http:// or https://, it's a URL so we treat it as an outgoing request
    if (StartsWithCaseInsensitive(filename, "http://") ||
        StartsWithCaseInsensitive(filename, "https://")) {
        eventId = EVENT_PRE_OUTGOING_REQUEST;
        eventCache.outgoingRequestUrl = filename;
    } else {
        eventId = EVENT_PRE_PATH_ACCESSED;
        eventCache.filename = filename;
    }
}

/* Helper for handle post file path access */
void helper_handle_post_file_path_access(EVENT_ID &eventId) {
    if (!eventCache.outgoingRequestUrl.empty()) {
        // If the pre handler for path access determined this was actually an URL,
        // we need to notify that the request finished.
        eventId = EVENT_POST_OUTGOING_REQUEST;

        // As we cannot extract the effective URL for these fopen wrappers,
        // we will just assume it's the same as the initial URL.
        eventCache.outgoingRequestEffectiveUrl = eventCache.outgoingRequestUrl;
    }
}

/* Handles PHP functions that have a file path as first parameter (pre-execution) */
AIKIDO_HANDLER_FUNCTION(handle_pre_file_path_access) {
    zend_string *filename = NULL;

    ZEND_PARSE_PARAMETERS_START(0, -1)
    Z_PARAM_OPTIONAL
    Z_PARAM_STR(filename)
    ZEND_PARSE_PARAMETERS_END();

    if (!filename) {
        return;
    }

    helper_handle_pre_file_path_access(ZSTR_VAL(filename), eventId);
}

/* Handles PHP functions that have a file path as first parameter (post-execution) */
AIKIDO_HANDLER_FUNCTION(handle_post_file_path_access) {
    helper_handle_post_file_path_access(eventId);
}

/* Handles PHP functions that have a file path as both first and second parameter (pre-execution) */
AIKIDO_HANDLER_FUNCTION(handle_pre_file_path_access_2) {
    zend_string *filename = NULL;
    zend_string *filename2 = NULL;

    ZEND_PARSE_PARAMETERS_START(0, -1)
    Z_PARAM_OPTIONAL
    Z_PARAM_STR(filename)
    Z_PARAM_STR(filename2)
    ZEND_PARSE_PARAMETERS_END();

    if (!filename) {
        return;
    }

    helper_handle_pre_file_path_access(ZSTR_VAL(filename), eventId);
    if (filename2) {
        eventCache.filename2 = ZSTR_VAL(filename2);
    }
}

/* Handles PHP functions that have a file path as first parameter (post-execution) */
AIKIDO_HANDLER_FUNCTION(handle_post_file_path_access_2) {
    helper_handle_post_file_path_access(eventId);
}