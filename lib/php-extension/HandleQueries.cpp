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
    auto& eventCacheStack = AIKIDO_GLOBAL(eventCacheStack);
    eventCacheStack.Top().moduleName = "PDO";
    eventCacheStack.Top().sqlQuery = std::string(ZSTR_VAL(query), ZSTR_LEN(query));
    eventCacheStack.Top().sqlDialect = GetSqlDialectFromPdo(pdo_object);
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
    auto& eventCacheStack = AIKIDO_GLOBAL(eventCacheStack);
    eventCacheStack.Top().moduleName = "PDO";
    eventCacheStack.Top().sqlQuery = std::string(ZSTR_VAL(query), ZSTR_LEN(query));
    eventCacheStack.Top().sqlDialect = GetSqlDialectFromPdo(pdo_object);
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

    if (!stmt->query_string) {
        return;
    }

    eventId = EVENT_PRE_SQL_QUERY_EXECUTED;
    auto& eventCacheStack = AIKIDO_GLOBAL(eventCacheStack);
    eventCacheStack.Top().moduleName = "PDOStatement";
    
#if PHP_VERSION_ID >= 80100
    eventCacheStack.Top().sqlQuery = std::string(ZSTR_VAL(stmt->query_string), ZSTR_LEN(stmt->query_string));
#else
    eventCacheStack.Top().sqlQuery = std::string((char*)stmt->query_string);
#endif

#if PHP_VERSION_ID >= 80500
    if (!stmt->database_object_handle) {
        eventCacheStack.Top().sqlDialect = "unknown";
        return;
    }
    eventCacheStack.Top().sqlDialect = GetSqlDialectFromPdo(stmt->database_object_handle);
#else
    eventCacheStack.Top().sqlDialect = GetSqlDialectFromPdo(&stmt->database_object_handle);
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
    auto& eventCacheStack = AIKIDO_GLOBAL(eventCacheStack);
    eventCacheStack.Top().moduleName = "mysqli";
    eventCacheStack.Top().sqlQuery = std::string(query, queryLength);
    eventCacheStack.Top().sqlDialect = "mysql";
}

AIKIDO_HANDLER_FUNCTION(handle_pre_pg_query) {
    scopedTimer.SetSink(sink, "sql_op");

    zval *firstArg = nullptr;
    zend_string *queryArg = nullptr;

    ZEND_PARSE_PARAMETERS_START(1, 2)
        Z_PARAM_ZVAL(firstArg)
        Z_PARAM_OPTIONAL
        Z_PARAM_STR(queryArg)
    ZEND_PARSE_PARAMETERS_END();

    zend_string *query = queryArg;
    if (ZEND_NUM_ARGS() == 1 && firstArg && Z_TYPE_P(firstArg) == IS_STRING) {
        query = Z_STR_P(firstArg);
    }

    if (!query) {
        return;
    }

    eventId = EVENT_PRE_SQL_QUERY_EXECUTED;
    auto& eventCacheStack = AIKIDO_GLOBAL(eventCacheStack);
    eventCacheStack.Top().moduleName = "pgsql";
    eventCacheStack.Top().sqlQuery = ZSTR_VAL(query);
    eventCacheStack.Top().sqlDialect = "postgres";
}

AIKIDO_HANDLER_FUNCTION(handle_pre_pg_query_params) {
    scopedTimer.SetSink(sink, "sql_op");

    zval *firstArg = nullptr;
    zval *secondArg = nullptr;
    zval *paramsArg = nullptr;

    ZEND_PARSE_PARAMETERS_START(2, 3)
        Z_PARAM_ZVAL(firstArg)
        Z_PARAM_ZVAL(secondArg)
        Z_PARAM_OPTIONAL
        Z_PARAM_ARRAY(paramsArg)
    ZEND_PARSE_PARAMETERS_END();

    zend_string *query = nullptr;
    if (ZEND_NUM_ARGS() == 2 && firstArg && Z_TYPE_P(firstArg) == IS_STRING) {
        query = Z_STR_P(firstArg);
    } else if (ZEND_NUM_ARGS() == 3 && secondArg && Z_TYPE_P(secondArg) == IS_STRING) {
        query = Z_STR_P(secondArg);
    }

    if (!query) {
        return;
    }

    eventId = EVENT_PRE_SQL_QUERY_EXECUTED;
    auto& eventCacheStack = AIKIDO_GLOBAL(eventCacheStack);
    eventCacheStack.Top().moduleName = "pgsql";
    eventCacheStack.Top().sqlQuery = ZSTR_VAL(query);
    eventCacheStack.Top().sqlDialect = "postgres";
}

AIKIDO_HANDLER_FUNCTION(handle_pre_pg_prepare) {
    scopedTimer.SetSink(sink, "sql_op");

    zval *firstArg = nullptr;
    zval *secondArg = nullptr;
    zend_string *queryArg = nullptr;

    ZEND_PARSE_PARAMETERS_START(2, 3)
        Z_PARAM_ZVAL(firstArg)
        Z_PARAM_ZVAL(secondArg)
        Z_PARAM_OPTIONAL
        Z_PARAM_STR(queryArg)
    ZEND_PARSE_PARAMETERS_END();

    zend_string *query = queryArg;
    if (ZEND_NUM_ARGS() == 2 && secondArg && Z_TYPE_P(secondArg) == IS_STRING) {
        query = Z_STR_P(secondArg);
    }

    if (!query) {
        return;
    }

    eventId = EVENT_PRE_SQL_QUERY_EXECUTED;
    auto& eventCacheStack = AIKIDO_GLOBAL(eventCacheStack);
    eventCacheStack.Top().moduleName = "pgsql";
    eventCacheStack.Top().sqlQuery = ZSTR_VAL(query);
    eventCacheStack.Top().sqlDialect = "postgres";
}
