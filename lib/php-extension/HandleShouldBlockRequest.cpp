#include "Includes.h"

zend_class_entry *blockingStatusClass = nullptr;

// This variable is used to check if auto_block_request function has already been called,
// in order to avoid multiple calls to this function.
bool checkedAutoBlock = false;

// This variable is used to check if should_block_request function has already been called,
// in order to avoid multiple calls to this function.
bool checkedShouldBlockRequest = false;

bool CheckBlocking(EVENT_ID eventId, bool& checkedBlocking) {
    if (checkedBlocking) {
        return true;
    }

    ScopedTimer scopedTimer("check_blocking", "aikido_op");

    try {
        std::string output;
        requestProcessor.SendEvent(eventId, output);
        action.Execute(output);
        checkedBlocking = true;
        return true;
    } catch (const std::exception &e) {
        AIKIDO_LOG_ERROR("Exception encountered in processing get blocking status event: %s\n", e.what());
    }
    return false;
}

ZEND_FUNCTION(should_block_request) {
    if (AIKIDO_GLOBAL(sapi_name) == "cli") {
        AIKIDO_LOG_DEBUG("should_block_request called in CLI mode! Skipping...\n");
        return;
    }

    if (AIKIDO_GLOBAL(disable) == true) {
        return;
    }

    if (!blockingStatusClass) {
        return;
    }

    requestProcessor.LoadConfigOnce();

    if (!CheckBlocking(EVENT_GET_BLOCKING_STATUS, checkedShouldBlockRequest)) {
        return;
    }

    object_init_ex(return_value, blockingStatusClass);
#if PHP_VERSION_ID >= 80000
    zend_object *obj = Z_OBJ_P(return_value);
    if (!obj) {
        return;
    }
#else
    zval *obj = return_value;
#endif
    zend_update_property_bool(blockingStatusClass, obj, "block", sizeof("block") - 1, action.Block());
    zend_update_property_string(blockingStatusClass, obj, "type", sizeof("type") - 1, action.Type());
    zend_update_property_string(blockingStatusClass, obj, "trigger", sizeof("trigger") - 1, action.Trigger());
    zend_update_property_string(blockingStatusClass, obj, "description", sizeof("description") - 1, action.Description());
    zend_update_property_string(blockingStatusClass, obj, "ip", sizeof("ip") - 1, action.Ip());
    zend_update_property_string(blockingStatusClass, obj, "user_agent", sizeof("user_agent") - 1, action.UserAgent());
}

ZEND_FUNCTION(auto_block_request) {
    if (AIKIDO_GLOBAL(sapi_name) == "cli") {
        AIKIDO_LOG_DEBUG("auto_block_request called in CLI mode! Skipping...\n");
        return;
    }

    if (AIKIDO_GLOBAL(disable) == true) {
        return;
    }

    requestProcessor.LoadConfigOnce();

    CheckBlocking(EVENT_GET_AUTO_BLOCKING_STATUS, checkedAutoBlock);
}

void RegisterAikidoBlockRequestStatusClass() {
    zend_class_entry ce;
    INIT_CLASS_ENTRY(ce, "AikidoBlockRequestStatus", NULL);  // Register class without methods
    blockingStatusClass = zend_register_internal_class(&ce);

    zend_declare_property_bool(blockingStatusClass, "block", sizeof("block") - 1, 0, ZEND_ACC_PUBLIC);
    zend_declare_property_string(blockingStatusClass, "type", sizeof("type") - 1, "", ZEND_ACC_PUBLIC);
    zend_declare_property_string(blockingStatusClass, "trigger", sizeof("trigger") - 1, "", ZEND_ACC_PUBLIC);
    zend_declare_property_string(blockingStatusClass, "description", sizeof("description") - 1, "", ZEND_ACC_PUBLIC);
    zend_declare_property_string(blockingStatusClass, "ip", sizeof("ip") - 1, "", ZEND_ACC_PUBLIC);
    zend_declare_property_string(blockingStatusClass, "user_agent", sizeof("user_agent") - 1, "", ZEND_ACC_PUBLIC);
}
