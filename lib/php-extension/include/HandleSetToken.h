#include "Includes.h"

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_aikido_set_token, 0, 1, _IS_BOOL, 0)
    ZEND_ARG_TYPE_INFO(0, token, IS_STRING, 0)
ZEND_END_ARG_INFO()

ZEND_FUNCTION(set_token);
