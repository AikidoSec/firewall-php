#include "Includes.h"

AIKIDO_HANDLER_FUNCTION(handle_pre_pdo_query) {
    zend_string *query = NULL;

    ZEND_PARSE_PARAMETERS_START(0, -1)
    Z_PARAM_OPTIONAL
    Z_PARAM_STR(query)
    ZEND_PARSE_PARAMETERS_END();

    if (!query) {
        return;
    }

    /*
            Get the current pdo object for which the query function was called, using the "getThis" PHP helper function.
            https://github.com/php/php-src/blob/5dd8bb0fa884efba40117a83d198f3847922c0a3/Zend/zend_API.h#L526
    */
    zval *pdo_object = getThis();
    if (!pdo_object) {
        return;
    }

    eventId = EVENT_PRE_SQL_QUERY_EXECUTED;
    eventCache.moduleName = "PDO";
    eventCache.sqlQuery = ZSTR_VAL(query);
    eventCache.sqlDialect = "unknown";

    zval retval;
    if (aikido_call_user_function_one_param("getAttribute", PDO_ATTR_DRIVER_NAME, &retval, pdo_object)) {
        if (Z_TYPE(retval) == IS_STRING) {
            eventCache.sqlDialect = Z_STRVAL_P(&retval);
        }
    }

    zval_ptr_dtor(&retval);
}
