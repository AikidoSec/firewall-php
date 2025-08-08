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


zend_class_entry* mysqli_link_class_entry = nullptr;

AIKIDO_HANDLER_FUNCTION(handle_pre_mysqli_query){
	zval				*mysql_link;
	char				*query = NULL;
	size_t 				query_len;
    zend_long 		    resultmode;

    if ((mysqli_link_class_entry = (zend_class_entry*)zend_hash_str_find_ptr(EG(class_table), "mysqli", sizeof("mysqli") - 1)) == NULL) {
        AIKIDO_LOG_DEBUG("handle_pre_mysqli_query: no link class\n");
        return;
    }

	if (zend_parse_method_parameters(ZEND_NUM_ARGS(), getThis(), "Os|l", &mysql_link, mysqli_link_class_entry, &query, &query_len, &resultmode) == FAILURE) {
		AIKIDO_LOG_DEBUG("handle_pre_mysqli_query: no parse\n");
        RETURN_THROWS();
	}

	if (!query_len) {
        AIKIDO_LOG_DEBUG("handle_pre_mysqli_query: no query len\n");
		RETURN_THROWS();
	}

    if (!mysql_link) {
        AIKIDO_LOG_WARN("handle_pre_mysqli_query: Missing mysql_object or query parameter.\n");
        return;
    }

    scopedTimer.SetSink(sink, "sql_op");

    eventId = EVENT_PRE_SQL_QUERY_EXECUTED;
    eventCache.moduleName = "mysqli";
    eventCache.sqlQuery = query;
    eventCache.sqlDialect = "mysql";
}
