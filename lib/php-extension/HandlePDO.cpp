#include "Includes.h"

AIKIDO_HANDLER_FUNCTION(handle_pre_pdo_query) {
    scopedTimer.SetSink(sink, "sql_op");

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
    eventCache.sqlDialect = GetSqlDialectFromPdo(pdo_object);
}

AIKIDO_HANDLER_FUNCTION(handle_pre_pdo_exec) {
    scopedTimer.SetSink(sink, "sql_op");
    zend_string *query = NULL;

    ZEND_PARSE_PARAMETERS_START(1, 1)
    Z_PARAM_STR(query)
    ZEND_PARSE_PARAMETERS_END();

    if (!query) {
        return;
    }

    zval *pdo_object = getThis();
    if (!pdo_object) {
        return;
    }

    eventId = EVENT_PRE_SQL_QUERY_EXECUTED;
    eventCache.moduleName = "PDO";
    eventCache.sqlQuery = ZSTR_VAL(query);
    eventCache.sqlDialect = GetSqlDialectFromPdo(pdo_object);
}

AIKIDO_HANDLER_FUNCTION(handle_pre_pdostatement_execute) {
    scopedTimer.SetSink(sink, "sql_op");

    zval *pdo_statement_object = getThis();
    if (!pdo_statement_object) {
        return;
    }

    pdo_stmt_t *stmt = Z_PDO_STMT_P(pdo_statement_object);
    if (!stmt->dbh) { // object is not initialized 
        return;
    }

    eventId = EVENT_PRE_SQL_QUERY_EXECUTED;
    eventCache.moduleName = "PDOStatement"; 
    eventCache.sqlQuery = PHP_GET_CHAR_PTR(stmt->query_string);    

    zval *pdo_object = &stmt->database_object_handle;
    eventCache.sqlDialect = GetSqlDialectFromPdo(pdo_object);
}