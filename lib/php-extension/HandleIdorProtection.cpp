#include "Includes.h"

// zval_get_string() on an array or object without __toString() emits a PHP
// warning or throws — reject those before calling it instead of letting that
// surface as a confusing error from inside this call.
static bool IsScalarZval(zval* value) {
    switch (Z_TYPE_P(value)) {
        case IS_STRING:
        case IS_LONG:
        case IS_DOUBLE:
        case IS_TRUE:
        case IS_FALSE:
            return true;
        default:
            return false;
    }
}

// Invalid UTF-8 would otherwise get silently mangled when this value is later
// JSON-encoded for the Go side, which would make it stop matching real column/
// table/tenant values forever (e.g. every query would look like it's missing
// its tenant filter). Reject it here instead, with a clear error.
static bool IsValidUtf8(const std::string& value) {
    try {
        json(value).dump();
        return true;
    } catch (const std::exception& e) {
        return false;
    }
}

ZEND_FUNCTION(enable_idor_protection) {
    if (IsAikidoDisabledOrBypassed()) {
        RETURN_BOOL(false);
    }

    char* tenantColumnName = nullptr;
    size_t tenantColumnNameLength = 0;
    zval* excludedTablesArr = nullptr;

    ZEND_PARSE_PARAMETERS_START(1, 2)
        Z_PARAM_STRING(tenantColumnName, tenantColumnNameLength)
        Z_PARAM_OPTIONAL
        Z_PARAM_ARRAY(excludedTablesArr)
    ZEND_PARSE_PARAMETERS_END();

    if (!tenantColumnName || tenantColumnNameLength == 0) {
        AIKIDO_LOG_ERROR("enable_idor_protection: tenantColumnName is null or empty!\n");
        RETURN_BOOL(false);
    }

    std::string tenantColumnNameStr(tenantColumnName, tenantColumnNameLength);
    if (!IsValidUtf8(tenantColumnNameStr)) {
        zend_throw_exception(GetFirewallDefaultExceptionCe(), "aikido\\enable_idor_protection(): tenantColumnName must be valid UTF-8", 0);
        RETURN_BOOL(false);
    }

    std::vector<std::string> excludedTables;
    if (excludedTablesArr) {
        zval* entry;
        ZEND_HASH_FOREACH_VAL(Z_ARRVAL_P(excludedTablesArr), entry) {
            if (!IsScalarZval(entry)) {
                AIKIDO_LOG_WARN("enable_idor_protection: skipping non-string entry in excludedTables!\n");
                continue;
            }
            zend_string* entryStr = zval_get_string(entry);
            std::string tableName(ZSTR_VAL(entryStr), ZSTR_LEN(entryStr));
            zend_string_release(entryStr);

            if (!IsValidUtf8(tableName)) {
                zend_throw_exception(GetFirewallDefaultExceptionCe(), "aikido\\enable_idor_protection(): excludedTables entries must be valid UTF-8", 0);
                RETURN_BOOL(false);
            }
            excludedTables.push_back(tableName);
        } ZEND_HASH_FOREACH_END();
    }

    auto& requestCache = AIKIDO_GLOBAL(requestCache);
    requestCache.idorTenantColumnName = tenantColumnNameStr;
    requestCache.idorExcludedTables = excludedTables;
    requestCache.idorProtectionEnabled = true;

    AIKIDO_LOG_DEBUG("Enabled IDOR protection with tenant column \"%s\" (%zu excluded tables)\n",
                      requestCache.idorTenantColumnName.c_str(), requestCache.idorExcludedTables.size());

    RETURN_BOOL(true);
}

ZEND_FUNCTION(set_tenant_id) {
    if (IsAikidoDisabledOrBypassed()) {
        return;
    }

    zval* tenantIdZv = nullptr;

    ZEND_PARSE_PARAMETERS_START(1, 1)
        Z_PARAM_ZVAL(tenantIdZv)
    ZEND_PARSE_PARAMETERS_END();

    if (!IsScalarZval(tenantIdZv)) {
        php_error_docref(NULL, E_WARNING, "aikido\\set_tenant_id(): tenantId must be a string or int");
        return;
    }

    zend_string* tenantIdStr = zval_get_string(tenantIdZv);
    if (!tenantIdStr || ZSTR_LEN(tenantIdStr) == 0) {
        if (tenantIdStr) {
            zend_string_release(tenantIdStr);
        }
        php_error_docref(NULL, E_WARNING, "aikido\\set_tenant_id(): tenantId must not be empty");
        return;
    }

    std::string tenantIdValue(ZSTR_VAL(tenantIdStr), ZSTR_LEN(tenantIdStr));
    zend_string_release(tenantIdStr);

    if (!IsValidUtf8(tenantIdValue)) {
        zend_throw_exception(GetFirewallDefaultExceptionCe(), "aikido\\set_tenant_id(): tenantId must be valid UTF-8", 0);
        return;
    }

    auto& requestCache = AIKIDO_GLOBAL(requestCache);
    requestCache.tenantId = tenantIdValue;
    requestCache.tenantIdSet = true;

    AIKIDO_LOG_DEBUG("Set tenant id to %s\n", requestCache.tenantId.c_str());
}

ZEND_FUNCTION(get_tenant_id) {
    ZEND_PARSE_PARAMETERS_NONE();

    auto& requestCache = AIKIDO_GLOBAL(requestCache);
    if (!requestCache.tenantIdSet) {
        RETURN_NULL();
    }

    RETURN_STRINGL(requestCache.tenantId.c_str(), requestCache.tenantId.length());
}

// PHP exceptions don't unwind the C++ stack, so the decrement below always runs.
ZEND_FUNCTION(without_idor_protection) {
    zend_fcall_info fci = {};
    zend_fcall_info_cache fcc = {};

    ZEND_PARSE_PARAMETERS_START(1, 1)
        Z_PARAM_FUNC(fci, fcc)
    ZEND_PARSE_PARAMETERS_END();

    auto& requestCache = AIKIDO_GLOBAL(requestCache);
    requestCache.EnterIdorIgnoredScope();

    zval retval;
    ZVAL_UNDEF(&retval);
    fci.retval = &retval;
    fci.param_count = 0;
    fci.params = nullptr;

    zend_call_function(&fci, &fcc);

    requestCache.LeaveIdorIgnoredScope();

    if (!Z_ISUNDEF(retval)) {
        RETVAL_ZVAL(&retval, 1, 1);
    }
}
