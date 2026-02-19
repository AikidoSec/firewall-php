#include "Includes.h"

zend_op_array* handle_file_compilation(zend_file_handle* file_handle, int type) {
    if(AIKIDO_GLOBAL(disable) == true) {
        return original_file_compilation_handler(file_handle, type);
    }

    auto& eventCacheStack = AIKIDO_GLOBAL(eventCacheStack);
    
    // Create a new event context for file compilation
    ScopedEventContext scopedContext;
    
    switch (type) {
        case ZEND_INCLUDE:
            eventCacheStack.Top().functionName = "include";
            break;
        case ZEND_INCLUDE_ONCE:
            eventCacheStack.Top().functionName = "include_once";
            break;
        case ZEND_REQUIRE:
            eventCacheStack.Top().functionName = "require";
            break;
        case ZEND_REQUIRE_ONCE:
            eventCacheStack.Top().functionName = "require_once";
            break;
        default:
            return original_file_compilation_handler(file_handle, type);
    }

    ScopedTimer scopedTimer(eventCacheStack.Top().functionName, "fs_op");

    char* filename = PHP_GET_CHAR_PTR(file_handle->filename);
    
    AIKIDO_LOG_DEBUG("\"%s\" called for \"%s\"!\n", eventCacheStack.Top().functionName.c_str(), filename);

    EVENT_ID eventId = NO_EVENT_ID;
    helper_handle_pre_file_path_access(filename, eventId);

    if (aikido_process_event(eventId, eventCacheStack.Top().functionName) == BLOCK) {
        // exit zend_compile_file handler and do not call the original handler, thus blocking the script file compilation
        return nullptr;
    }

    scopedTimer.Stop();
    zend_op_array* op_array = original_file_compilation_handler(file_handle, type);
    scopedTimer.Start();

    eventId = NO_EVENT_ID;
    helper_handle_post_file_path_access(eventId);
    aikido_process_event(eventId, eventCacheStack.Top().functionName);

    return op_array;
}
