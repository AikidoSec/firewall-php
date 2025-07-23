#include "Includes.h"

std::string get_resource_or_original_from_php_filter(const std::string& filenameStr) {
    std::string phpResourceString = "/resource=";
    size_t pos = filenameStr.rfind(phpResourceString);
    if (pos != std::string::npos) {
        return filenameStr.substr(pos + phpResourceString.length());
    }
    return filenameStr;
}

/* Helper for handle pre file path access */
void helper_handle_pre_file_path_access(char *filename, EVENT_ID &eventId) {
    std::string filenameString(filename);

    //https://github.com/php/php-src/blob/8b61c49987750b74bee19838c7f7c9fbbf53aace/ext/standard/php_fopen_wrapper.c#L339
    if (StartsWith(filename, "php://", false) && !StartsWith(filename, "php://filter", false)) {
        // Whitelist all php:// streams apart from php://filter, for performance reasons (some PHP frameworks do 1000+ calls / request with these streams as param)
        // php://filter can be used to open arbitrary files, so we still monitor this
        return;
    }

    filenameString = get_resource_or_original_from_php_filter(filenameString);

    // if filename starts with http:// or https://, it's a URL so we treat it as an outgoing request
    if (StartsWith(filenameString, "http://", false) ||
        StartsWith(filenameString, "https://", false)) {
        eventId = EVENT_PRE_OUTGOING_REQUEST;
        eventCache.outgoingRequestUrl = filenameString;
    } else {
        eventId = EVENT_PRE_PATH_ACCESSED;
        eventCache.filename = filenameString;
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
    scopedTimer.SetSink(sink, "fs_op");

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
    scopedTimer.SetSink(sink, "fs_op");

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