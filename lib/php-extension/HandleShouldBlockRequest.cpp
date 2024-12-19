#include "Includes.h"

zend_class_entry *blockingStatusClass = nullptr;

bool GetBlockingStatus() {
    try {
        std::string output;
        requestProcessor.SendEvent(EVENT_GET_BLOCKING_STATUS, output);
        action.Execute(output);
        return true;
    } catch (const std::exception &e) {
        AIKIDO_LOG_ERROR("Exception encountered in processing get blocking status event: %s\n", e.what());
    }
    return false;
}

ZEND_FUNCTION(should_block_request) {
    ScopedTimer scopedTimer("should_block_request");

    if (AIKIDO_GLOBAL(disable) == true) {
        return;
    }

    if (!blockingStatusClass) {
        return;
    }

    requestProcessor.LoadConfigOnce();

    if (!GetBlockingStatus()) {
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
}
