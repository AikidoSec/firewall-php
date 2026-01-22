#include "Includes.h"

void helper_handle_pre_shell_execution(std::string cmd, EVENT_ID &eventId) {
    eventCacheStack.Top().cmd = cmd;
    eventId = EVENT_PRE_SHELL_EXECUTED;
}

AIKIDO_HANDLER_FUNCTION(handle_shell_execution) {
    scopedTimer.SetSink(sink, "exec_op");

    zend_string *cmd = NULL;

    ZEND_PARSE_PARAMETERS_START(0, -1)
    Z_PARAM_OPTIONAL
    Z_PARAM_STR(cmd)
    ZEND_PARSE_PARAMETERS_END();

    if (!cmd) {
        return;
    }

    helper_handle_pre_shell_execution(ZSTR_VAL(cmd), eventId);
}


AIKIDO_HANDLER_FUNCTION(handle_shell_execution_with_array) {
    zval *cmdVal = nullptr;

    ZEND_PARSE_PARAMETERS_START(0, -1)
    Z_PARAM_OPTIONAL
    Z_PARAM_ZVAL(cmdVal)
    ZEND_PARSE_PARAMETERS_END();

    if (Z_TYPE_P(cmdVal) == IS_STRING) {
        zend_string* cmdStr = Z_STR_P(cmdVal);
        if (!cmdStr) {
            return;
        }
        helper_handle_pre_shell_execution(ZSTR_VAL(cmdStr), eventId);
    }
}
