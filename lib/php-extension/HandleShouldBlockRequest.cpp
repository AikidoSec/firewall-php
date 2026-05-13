#include "Includes.h"

zend_class_entry *blockingStatusClass = nullptr;
zend_class_entry *whitelistStatusClass = nullptr;

bool CheckBlocking(EVENT_ID eventId, bool& checkedBlocking) {
    if (checkedBlocking) {
        return true;
    }

    ScopedTimer scopedTimer("check_blocking", "aikido_op");

    try {
        auto& requestProcessorInstance = AIKIDO_GLOBAL(requestProcessorInstance);
        auto& action = AIKIDO_GLOBAL(action);
        std::string output;
        requestProcessorInstance.SendEvent(eventId, output);
        action.Execute(output);
        checkedBlocking = true;
        return true;
    } catch (const std::exception &e) {
        AIKIDO_LOG_ERROR("Exception encountered in processing get blocking status event: %s\n", e.what());
    }
    return false;
}

bool CheckWhitelist(EVENT_ID eventId, bool& checkedWhitelist) {
    if (checkedWhitelist) {
        return true;
    }

    ScopedTimer scopedTimer("check_whitelist", "aikido_op");

    try {
        auto& requestProcessorInstance = AIKIDO_GLOBAL(requestProcessorInstance);
        auto& action = AIKIDO_GLOBAL(action);
        std::string output;
        requestProcessorInstance.SendEvent(eventId, output);
        action.Execute(output);
        checkedWhitelist = true;
        return true;
    } catch (const std::exception &e) {
        AIKIDO_LOG_ERROR("Exception encountered in processing get whitelist status event: %s\n", e.what());
    }
    return false;
}

ZEND_FUNCTION(should_block_request) {
    if (AIKIDO_GLOBAL(sapi_name) == "cli") {
        AIKIDO_LOG_DEBUG("should_block_request called in CLI mode! Skipping...\n");
        return;
    }

    if (!blockingStatusClass) {
        return;
    }

    // Initialize the returned object with default values so that block = false
    // even if the IP is bypassed
    object_init_ex(return_value, blockingStatusClass);

    if (IsAikidoDisabledOrBypassed()) {
        return;
    }

    if (!CheckBlocking(EVENT_GET_BLOCKING_STATUS, AIKIDO_GLOBAL(checkedShouldBlockRequest))) {
        return;
    }

#if PHP_VERSION_ID >= 80000
    zend_object *obj = Z_OBJ_P(return_value);
    if (!obj) {
        return;
    }
#else
    zval *obj = return_value;
#endif
    auto& action = AIKIDO_GLOBAL(action);
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

    if (IsAikidoDisabledOrBypassed()) {
        return;
    }

    CheckBlocking(EVENT_GET_AUTO_BLOCKING_STATUS, AIKIDO_GLOBAL(checkedAutoBlock));
}

ZEND_FUNCTION(should_whitelist_request) {
    if (AIKIDO_GLOBAL(sapi_name) == "cli") {
        AIKIDO_LOG_DEBUG("should_whitelist_request called in CLI mode! Skipping...\n");
        return;
    }

    if (!whitelistStatusClass) {
        return;
    }

    if (IsAikidoDisabled()) {
        return;
    }

    object_init_ex(return_value, whitelistStatusClass);

    if (!CheckWhitelist(EVENT_GET_WHITELISTED_STATUS, AIKIDO_GLOBAL(checkedWhitelistRequest))) {
        return;
    }

    #if PHP_VERSION_ID >= 80000
        zend_object *obj = Z_OBJ_P(return_value);
        if (!obj) {
            return;
        }
    #else
        zval *obj = return_value;
    #endif

    auto& action = AIKIDO_GLOBAL(action);
    zend_update_property_bool(whitelistStatusClass, obj, "whitelisted", sizeof("whitelisted") - 1, action.Whitelisted());
    zend_update_property_string(whitelistStatusClass, obj, "type", sizeof("type") - 1, action.Type());
    zend_update_property_string(whitelistStatusClass, obj, "trigger", sizeof("trigger") - 1, action.Trigger());
    zend_update_property_string(whitelistStatusClass, obj, "description", sizeof("description") - 1, action.Description());
    zend_update_property_string(whitelistStatusClass, obj, "ip", sizeof("ip") - 1, action.Ip());
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

void RegisterAikidoWhitelistRequestStatusClass() {
    zend_class_entry ce;
    INIT_CLASS_ENTRY(ce, "AikidoWhitelistRequestStatus", NULL);  // Register class without methods
    whitelistStatusClass = zend_register_internal_class(&ce);

    zend_declare_property_bool(whitelistStatusClass, "whitelisted", sizeof("whitelisted") - 1, 0, ZEND_ACC_PUBLIC);
    zend_declare_property_string(whitelistStatusClass, "type", sizeof("type") - 1, "", ZEND_ACC_PUBLIC);
    zend_declare_property_string(whitelistStatusClass, "trigger", sizeof("trigger") - 1, "", ZEND_ACC_PUBLIC);
    zend_declare_property_string(whitelistStatusClass, "description", sizeof("description") - 1, "", ZEND_ACC_PUBLIC);
    zend_declare_property_string(whitelistStatusClass, "ip", sizeof("ip") - 1, "", ZEND_ACC_PUBLIC);
}
