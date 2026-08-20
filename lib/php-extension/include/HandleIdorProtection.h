#pragma once

ZEND_FUNCTION(enable_idor_protection);
ZEND_FUNCTION(set_tenant_id);
ZEND_FUNCTION(get_tenant_id);
ZEND_FUNCTION(without_idor_protection);

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_aikido_enable_idor_protection, 0, 1, _IS_BOOL, 0)
    ZEND_ARG_TYPE_INFO(0, tenantColumnName, IS_STRING, 0)
    ZEND_ARG_TYPE_INFO(0, excludedTables, IS_ARRAY, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_aikido_set_tenant_id, 0, 1, IS_VOID, 0)
    ZEND_ARG_INFO(0, tenantId)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_aikido_get_tenant_id, 0, 0, IS_STRING, 1)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_aikido_without_idor_protection, 0, 1, IS_MIXED, 0)
    ZEND_ARG_TYPE_INFO(0, fn, IS_CALLABLE, 0)
ZEND_END_ARG_INFO()
