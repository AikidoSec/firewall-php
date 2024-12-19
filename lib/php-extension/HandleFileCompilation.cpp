#include "Includes.h"

zend_op_array* handle_file_compilation(zend_file_handle* file_handle, int type) {
    eventCache.Reset();
    switch (type) {
        case ZEND_INCLUDE:
            eventCache.functionName = "include";
            break;
        case ZEND_INCLUDE_ONCE:
            eventCache.functionName = "include_once";
            break;
        case ZEND_REQUIRE:
            eventCache.functionName = "require";
            break;
        case ZEND_REQUIRE_ONCE:
            eventCache.functionName = "require_once";
            break;
        default:
            return original_file_compilation_handler(file_handle, type);
    }

    #if PHP_VERSION_ID >= 80100
    char* filename = ZSTR_VAL(file_handle->filename);
    #else
    char* filename = (char*)file_handle->filename;
    #endif
    
    AIKIDO_LOG_DEBUG("\"%s\" called for \"%s\"!\n", eventCache.functionName.c_str(), filename);

    EVENT_ID eventId = NO_EVENT_ID;
    helper_handle_pre_file_path_access(filename, eventId);

    if (aikido_process_event(eventId) == BLOCK) {
        // exit zend_compile_file handler and do not call the original handler, thus blocking the script file compilation
        return nullptr;
    }

    zend_op_array* op_array = original_file_compilation_handler(file_handle, type);

    eventId = NO_EVENT_ID;
    helper_handle_post_file_path_access(eventId);
    aikido_process_event(eventId);

    return op_array;
}
