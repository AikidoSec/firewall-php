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
    eventCacheStack.Top().moduleName = "PDO";
    eventCacheStack.Top().sqlQuery = ZSTR_VAL(query);
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
    eventCacheStack.Top().moduleName = "PDO";
    eventCacheStack.Top().sqlQuery = ZSTR_VAL(query);
    eventCacheStack.Top().sqlDialect = GetSqlDialectFromPdo(pdo_object);
}

/*
    bindParam() binds a variable by reference — PHP stores this as an IS_REFERENCE
    zval wrapping the actual value. Without unwrapping, Z_TYPE_P would return
    IS_REFERENCE instead of the real type (IS_STRING, IS_LONG, ...) and the switch
    below would fail to extract the value. bindValue() stores values directly, so
    no unwrapping is needed for those.
*/
static bool ZvalToString(zval *entry, std::string &out) {
    if (Z_TYPE_P(entry) == IS_REFERENCE) {
        entry = Z_REFVAL_P(entry);
    }

    switch (Z_TYPE_P(entry)) {
        case IS_STRING:
            out = std::string(Z_STRVAL_P(entry), Z_STRLEN_P(entry));
            return true;
        case IS_LONG:
            out = std::to_string(Z_LVAL_P(entry));
            return true;
        case IS_DOUBLE:
            out = std::to_string(Z_DVAL_P(entry));
            return true;
        case IS_TRUE:
            out = "1";
            return true;
        case IS_FALSE:
            out = "0";
            return true;
        case IS_NULL:
            out = "";
            return true;
        default:
            return false;
    }
}

/* Unhandled types (arrays, objects) are stored as null to preserve positional indices. */
static std::string ConvertParamsToJson(zval *params) {
    if (!params || Z_TYPE_P(params) != IS_ARRAY) {
        return "";
    }

    HashTable *ht = Z_ARRVAL_P(params);
    json paramsJson;

    bool isIndexed = true;
    zend_string *key;
    zend_ulong idx;
    ZEND_HASH_FOREACH_KEY(ht, idx, key) {
        if (key) {
            isIndexed = false;
            break;
        }
        (void)idx;
    } ZEND_HASH_FOREACH_END();

    if (isIndexed) {
        paramsJson = json::array();
    } else {
        paramsJson = json::object();
    }

    zval *entry;
    ZEND_HASH_FOREACH_KEY_VAL(ht, idx, key, entry) {
        std::string valueStr;
        bool resolved = ZvalToString(entry, valueStr);

        if (isIndexed) {
            if (resolved) {
                paramsJson.push_back(valueStr);
            } else {
                paramsJson.push_back(nullptr);
            }
        } else if (key && resolved) {
            paramsJson[std::string(ZSTR_VAL(key), ZSTR_LEN(key))] = valueStr;
        }
        (void)idx;
    } ZEND_HASH_FOREACH_END();

    return paramsJson.dump();
}

/* Fallback for when execute() is called without inline params (bindValue/bindParam). */
static std::string ConvertBoundParamsToJson(HashTable *bound_params) {
    if (!bound_params || zend_hash_num_elements(bound_params) == 0) {
        return "";
    }

    bool hasNamed = false;
    zend_string *key;
    zend_ulong idx;
    ZEND_HASH_FOREACH_KEY(bound_params, idx, key) {
        if (key) {
            hasNamed = true;
            break;
        }
        (void)idx;
    } ZEND_HASH_FOREACH_END();

    json paramsJson;
    if (hasNamed) {
        paramsJson = json::object();
    } else {
        paramsJson = json::array();
    }

    zval *bound_zval;
    ZEND_HASH_FOREACH_VAL(bound_params, bound_zval) {
        auto *param = (struct pdo_bound_param_data *)Z_PTR_P(bound_zval);
        std::string valueStr;
        bool resolved = ZvalToString(&param->parameter, valueStr);

        if (hasNamed && param->name) {
            if (resolved) {
                paramsJson[std::string(ZSTR_VAL(param->name), ZSTR_LEN(param->name))] = valueStr;
            }
        } else {
            if (resolved) {
                paramsJson.push_back(valueStr);
            } else {
                paramsJson.push_back(nullptr);
            }
        }
    } ZEND_HASH_FOREACH_END();

    return paramsJson.dump();
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

    zval *inputParams = NULL;
    ZEND_PARSE_PARAMETERS_START(0, 1)
        Z_PARAM_OPTIONAL
        Z_PARAM_ARRAY(inputParams)
    ZEND_PARSE_PARAMETERS_END();

    eventId = EVENT_PRE_SQL_QUERY_EXECUTED;
    eventCacheStack.Top().moduleName = "PDOStatement";
    eventCacheStack.Top().sqlQuery = PHP_GET_CHAR_PTR(stmt->query_string);

    std::string sqlParams = ConvertParamsToJson(inputParams);
    if (sqlParams.empty() && stmt->bound_params) {
        sqlParams = ConvertBoundParamsToJson(stmt->bound_params);
    }
    eventCacheStack.Top().sqlParams = sqlParams;

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
    eventCacheStack.Top().moduleName = "mysqli";
    eventCacheStack.Top().sqlQuery = query;
    eventCacheStack.Top().sqlDialect = "mysql";
}
