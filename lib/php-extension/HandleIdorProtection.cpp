#include "Includes.h"

ZEND_FUNCTION(enable_idor_protection) {
    ScopedTimer scopedTimer("enable_idor_protection", "aikido_op");
    
    if (IsAikidoDisabledOrBypassed()) {
        RETURN_BOOL(false);
    }

    char *tenantColumnName = nullptr;
    size_t tenantColumnNameLength = 0;
    zval *excludedTablesZval = nullptr;

    ZEND_PARSE_PARAMETERS_START(1, 2)
        Z_PARAM_STRING(tenantColumnName, tenantColumnNameLength)
        Z_PARAM_OPTIONAL
        Z_PARAM_ARRAY(excludedTablesZval)
    ZEND_PARSE_PARAMETERS_END();

    if (!tenantColumnName || tenantColumnNameLength == 0) {
        AIKIDO_LOG_ERROR("enable_idor_protection: tenant_column_name is null or empty!\n");
        RETURN_BOOL(false);
    }

    json excludedTablesJson = json::array();
    if (excludedTablesZval) {
        HashTable *ht = Z_ARRVAL_P(excludedTablesZval);
        zval *entry;
        ZEND_HASH_FOREACH_VAL(ht, entry) {
            if (Z_TYPE_P(entry) == IS_STRING) {
                excludedTablesJson.push_back(std::string(Z_STRVAL_P(entry), Z_STRLEN_P(entry)));
            }
        } ZEND_HASH_FOREACH_END();
    }

    json idorConfig = {
        {"column_name", std::string(tenantColumnName, tenantColumnNameLength)},
        {"excluded_tables", excludedTablesJson}
    };
    requestCache.idorConfigJson = idorConfig.dump();

    AIKIDO_LOG_INFO("Enabled IDOR protection with tenant column '%s'\n", tenantColumnName);
    RETURN_BOOL(true);
}

ZEND_FUNCTION(set_tenant_id) {
    if (IsAikidoDisabledOrBypassed()) {
        return;
    }

    char *id = nullptr;
    size_t idLength = 0;

    ZEND_PARSE_PARAMETERS_START(1, 1)
        Z_PARAM_STRING(id, idLength)
    ZEND_PARSE_PARAMETERS_END();

    if (!id || idLength == 0) {
        AIKIDO_LOG_ERROR("set_tenant_id: id is null or empty!\n");
        return;
    }

    requestCache.tenantId = std::string(id, idLength);
    AIKIDO_LOG_DEBUG("Set tenant ID to %s\n", requestCache.tenantId.c_str());
}

ZEND_FUNCTION(without_idor_protection) {
    zend_fcall_info fci;
    zend_fcall_info_cache fci_cache;

    ZEND_PARSE_PARAMETERS_START(1, 1)
        Z_PARAM_FUNC(fci, fci_cache)
    ZEND_PARSE_PARAMETERS_END();

    requestCache.idorDisabled = true;

    zval retval;
    ZVAL_UNDEF(&retval);
    fci.retval = &retval;

    zend_result result = (zend_result)zend_call_function(&fci, &fci_cache);

    requestCache.idorDisabled = false;

    if (result == SUCCESS && !EG(exception)) {
        if (!Z_ISUNDEF(retval)) {
            ZVAL_COPY_VALUE(return_value, &retval);
        }
    } else {
        zval_ptr_dtor(&retval);
    }
}
