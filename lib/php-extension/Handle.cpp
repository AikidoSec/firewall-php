#include "Includes.h"
#include "include/Stats.h"

ACTION_STATUS aikido_process_event(EVENT_ID& eventId, std::string& sink) {
    if (eventId == NO_EVENT_ID) {
        return CONTINUE;
    }

    std::string outputEvent;
    requestProcessor.SendEvent(eventId, outputEvent);

    if (action.IsDetection(outputEvent)) {
        stats[sink].IncrementAttacksDetected();
    }

    if (!requestProcessor.IsBlockingEnabled()) {
        return CONTINUE;
    }

    ACTION_STATUS action_status = action.Execute(outputEvent);
    if (action_status == BLOCK) {
        stats[sink].IncrementAttacksBlocked();
    }
    return action_status;
}

ZEND_NAMED_FUNCTION(aikido_generic_handler) {
    ScopedTimer scopedTimer;
    ScopedEventContext scopedContext; // RAII: auto push/pop event context

    AIKIDO_LOG_DEBUG("Aikido generic handler started!\n");

    zif_handler original_handler = nullptr;
    aikido_handler post_handler = nullptr;

    std::string sink;
    std::string outputEvent;
    bool caughtException = false;

    try {
        zend_execute_data* exec_data = EG(current_execute_data);
        zend_function* func = exec_data->func;
        zend_class_entry* executed_scope = exec_data->func->common.scope;

        std::string function_name(ZSTR_VAL(func->common.function_name));
        function_name = ToLowercase(function_name);

        aikido_handler handler = nullptr;

        std::string scope_name = function_name;
        AIKIDO_LOG_DEBUG("Function name: %s\n", scope_name.c_str());
        if (HOOKED_FUNCTIONS.find(function_name) != HOOKED_FUNCTIONS.end() && !executed_scope) {
            handler = HOOKED_FUNCTIONS[function_name].handler;
            post_handler = HOOKED_FUNCTIONS[function_name].post_handler;
            original_handler = HOOKED_FUNCTIONS[function_name].original_handler;
        } else if (executed_scope) {
            /* A method was executed (executed_scope stores the name of the current class) */

            std::string class_name(ZSTR_VAL(executed_scope->name));
            class_name = ToLowercase(class_name);

            scope_name = class_name + "->" + function_name;

            AIKIDO_METHOD_KEY method_key(class_name, function_name);

            AIKIDO_LOG_DEBUG("Method name: %s\n", scope_name.c_str());

            if (HOOKED_METHODS.find(method_key) == HOOKED_METHODS.end()) {
                AIKIDO_LOG_DEBUG("Method not found! Returning!\n");
                return;
            }

            handler = HOOKED_METHODS[method_key].handler;
            post_handler = HOOKED_METHODS[method_key].post_handler;
            original_handler = HOOKED_METHODS[method_key].original_handler;
        } else {
            AIKIDO_LOG_DEBUG("Nothing matches the current handler! Returning!\n");
            return;
        }

        if (IsAikidoDisabledOrBypassed()) {
            if (original_handler) {
                original_handler(INTERNAL_FUNCTION_PARAM_PASSTHRU);
            }
            if (AIKIDO_GLOBAL(disable) == true) {
                AIKIDO_LOG_INFO("Aikido generic handler finished earlier because AIKIDO_DISABLE is set to 1!\n");
            } else {
                AIKIDO_LOG_INFO("Aikido generic handler finished earlier because IP is bypassed!\n");
            }
            return;
        }

        eventCacheStack.Current().functionName = scope_name;
        sink = scope_name;

        AIKIDO_LOG_DEBUG("Calling handler for \"%s\"!\n", scope_name.c_str());

        EVENT_ID eventId = NO_EVENT_ID;
        /*
                The handler for a specific PHP function that we hook can set an event ID
                to be sent to the Go libary (request processor).
                This will notify the Go library that an event has happend in the PHP layer.
                The event ID is initialy empty and it's only sent to Go only if the C++ handler
                for the currently hooked function sets it.
        */
        handler(INTERNAL_FUNCTION_PARAM_PASSTHRU, eventId, sink, scopedTimer);

        if (aikido_process_event(eventId, sink) == BLOCK) {
            // exit generic handler and do not call the original handler, thus blocking the execution
            AIKIDO_LOG_DEBUG("Aikido generic handler ended (block)!\n");
            return;
        }
    } catch (const std::exception& e) {
        caughtException = true;
        AIKIDO_LOG_ERROR("Exception encountered in generic handler: %s\n", e.what());
    }

    if (original_handler) {
        scopedTimer.Stop();
        original_handler(INTERNAL_FUNCTION_PARAM_PASSTHRU);
        scopedTimer.Start();

        if (!caughtException && post_handler) {
            EVENT_ID eventId = NO_EVENT_ID;
            /*
                    The handler for a specific PHP function that we hook can set an event ID
                    to be sent to the Go libary (request processor).
                    This will notify the Go library that an event has happend in the PHP layer.
                    The event ID is initialy empty and it's only sent to Go only if the C++ handler
                    for the currently hooked function sets it.
            */
            post_handler(INTERNAL_FUNCTION_PARAM_PASSTHRU, eventId, sink, scopedTimer);
            aikido_process_event(eventId, sink);
        }
    }

    AIKIDO_LOG_DEBUG("Aikido generic handler ended!\n");
}