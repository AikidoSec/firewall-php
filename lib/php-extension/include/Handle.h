#pragma once

/*
        Macro for registering an Aikido handler in the HOOKED_FUNCTIONS map.
        It takes as parameters the PHP function name to be hooked, a C++ function
        that should be called BEFORE that PHP function is executed and a C++ function
        that should be called AFTER that PHP function is executed.
        The nullptr part is a placeholder where the original function handler from
        the Zend framework will be stored at initialization when we run the hooking.
*/
#define AIKIDO_REGISTER_FUNCTION_HANDLER_WITH_POST_EX(function_name, pre_handler, post_handler) \
    {                                                                                           \
        std::string(#function_name), { pre_handler, post_handler, nullptr }                     \
    }

/*
        Macro for registering an Aikido handler in the HOOKED_FUNCTIONS map.
        It takes as parameters the PHP function name to be hooked and a C++ function
        that should be called BEFORE that PHP function is executed.
        This macro doesn't register any hook for AFTER the function is execute, that's why
        the second argument is nullptr.
        The last nullptr part is a placeholder where the original function handler from
        the Zend framework will be stored at initialization when we run the hooking.
*/
#define AIKIDO_REGISTER_FUNCTION_HANDLER_EX(function_name, pre_handler) AIKIDO_REGISTER_FUNCTION_HANDLER_WITH_POST_EX(function_name, pre_handler, nullptr)

/*
        Shorthand version of AIKIDO_REGISTER_FUNCTION_HANDLER_EX that constructs automatically the C++ function to be called.
        This version only registers a pre-hook (hook to be called before the original function is executed).
        For example, if function name is curl_init this macro will store { "curl_init", { handle_pre_curl_init, nullptr } }.
*/
#define AIKIDO_REGISTER_FUNCTION_HANDLER(function_name)                               \
    {                                                                                 \
        std::string(#function_name), { handle_pre_##function_name, nullptr, nullptr } \
    }

/*
        Shorthand version of AIKIDO_REGISTER_FUNCTION_HANDLER_WITH_POST_EX that constructs automatically the C++ function to be called.
        This version registers a pre-hook and a post-hook (hooks for before and after the function is executed).
        For example, if function name is curl_init this macro will store { "curl_init", { handle_pre_curl_init, handle_post_curl_init, nullptr } }.
*/
#define AIKIDO_REGISTER_FUNCTION_HANDLER_WITH_POST(function_name) AIKIDO_REGISTER_FUNCTION_HANDLER_WITH_POST_EX(function_name, handle_pre_##function_name, handle_post_##function_name)

/*
        Similar to AIKIDO_REGISTER_FUNCTION_HANDLER, but for methods.
*/
#define AIKIDO_REGISTER_METHOD_HANDLER(class_name, method_name)                                                                      \
    {                                                                                                                                \
        AIKIDO_METHOD_KEY(std::string(#class_name), std::string(#method_name)), { handle_pre_##class_name##_##method_name, nullptr } \
    }

#define AIKIDO_HANDLER_FUNCTION(name) void name(INTERNAL_FUNCTION_PARAMETERS, EVENT_ID& eventId)

ZEND_NAMED_FUNCTION(aikido_generic_handler);

ACTION_STATUS aikido_process_event(EVENT_ID& eventId);
