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
    auto& eventCache = AIKIDO_GLOBAL(eventCache);
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
    auto& eventCache = AIKIDO_GLOBAL(eventCache);
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
    auto& eventCache = AIKIDO_GLOBAL(eventCache);
    eventCache.moduleName = "PDOStatement"; 
    eventCache.sqlQuery = PHP_GET_CHAR_PTR(stmt->query_string);    

#if PHP_VERSION_ID >= 80500
    if (!stmt->database_object_handle) {
        eventCache.sqlDialect = "unknown";
        return;
    }
    eventCache.sqlDialect = GetSqlDialectFromPdo(stmt->database_object_handle);
#else
    eventCache.sqlDialect = GetSqlDialectFromPdo(&stmt->database_object_handle);
#endif
}

zend_class_entry* helper_load_mysqli_link_class_entry() {
    /* Static variable initialization ensures that the class entry is loaded only once and is thread-safe */
    static zend_class_entry* mysqliLinkClassEntry = (zend_class_entry*)zend_hash_str_find_ptr(EG(class_table), "mysqli", sizeof("mysqli") - 1);
    return mysqliLinkClassEntry;
}

AIKIDO_HANDLER_FUNCTION(handle_pre_mysqli_query){
	zval*     mysqliLinkObject = nullptr;
	char*	  query = nullptr;
	size_t 	  queryLength;
    zend_long resultMode;

    zend_class_entry* mysqliLinkClassEntry = helper_load_mysqli_link_class_entry();
    if (!mysqliLinkClassEntry) {
        AIKIDO_LOG_WARN("handle_pre_mysqli_query: did not find mysqli link class!\n");
        return;
    }

	if (zend_parse_method_parameters(ZEND_NUM_ARGS(), getThis(), "Os|l", &mysqliLinkObject, mysqliLinkClassEntry, &query, &queryLength, &resultMode) == FAILURE) {
		AIKIDO_LOG_WARN("handle_pre_mysqli_query: failed to parse parameters!\n");
        return;
	}

	if (!queryLength) {
        AIKIDO_LOG_WARN("handle_pre_mysqli_query: query length is 0!\n");
		return;
	}

    if (!mysqliLinkObject) {
        AIKIDO_LOG_WARN("handle_pre_mysqli_query: mysqli link object is null!\n");
        return;
    }

    scopedTimer.SetSink(sink, "sql_op");

    eventId = EVENT_PRE_SQL_QUERY_EXECUTED;
    auto& eventCache = AIKIDO_GLOBAL(eventCache);
    eventCache.moduleName = "mysqli";
    eventCache.sqlQuery = query;
    eventCache.sqlDialect = "mysql";
}
