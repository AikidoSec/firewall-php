#include "Includes.h"

zend_op_array* handle_file_compilation(zend_file_handle* file_handle, int type) {
    AIKIDO_GLOBAL(eventCache).Reset();
    switch (type) {
        case ZEND_INCLUDE:
            AIKIDO_GLOBAL(eventCache).functionName = "include";
            break;
        case ZEND_INCLUDE_ONCE:
            AIKIDO_GLOBAL(eventCache).functionName = "include_once";
            break;
        case ZEND_REQUIRE:
            AIKIDO_GLOBAL(eventCache).functionName = "require";
            break;
        case ZEND_REQUIRE_ONCE:
            AIKIDO_GLOBAL(eventCache).functionName = "require_once";
            break;
        default:
            return original_file_compilation_handler(file_handle, type);
    }

    ScopedTimer scopedTimer(AIKIDO_GLOBAL(eventCache).functionName, "fs_op");

    char* filename = PHP_GET_CHAR_PTR(file_handle->filename);
    
    AIKIDO_LOG_DEBUG("\"%s\" called for \"%s\"!\n", AIKIDO_GLOBAL(eventCache).functionName.c_str(), filename);

    EVENT_ID eventId = NO_EVENT_ID;
    helper_handle_pre_file_path_access(filename, eventId);

    if (aikido_process_event(eventId, AIKIDO_GLOBAL(eventCache).functionName) == BLOCK) {
        // exit zend_compile_file handler and do not call the original handler, thus blocking the script file compilation
        return nullptr;
    }

    scopedTimer.Stop();
    zend_op_array* op_array = original_file_compilation_handler(file_handle, type);
    scopedTimer.Start();

    eventId = NO_EVENT_ID;
    helper_handle_post_file_path_access(eventId);
    aikido_process_event(eventId, AIKIDO_GLOBAL(eventCache).functionName);

    return op_array;
}
