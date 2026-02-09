#pragma once

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_aikido_worker_rinit, 0, 0, _IS_BOOL, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_aikido_worker_rshutdown, 0, 0, _IS_BOOL, 0)
ZEND_END_ARG_INFO()

PHP_FUNCTION(worker_rinit);
PHP_FUNCTION(worker_rshutdown);
