#include "Includes.h"

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_aikido_register_param_matcher, 0, 2, _IS_BOOL, 0)
    ZEND_ARG_TYPE_INFO(0, param, IS_STRING, 0)
    ZEND_ARG_TYPE_INFO(0, regex, IS_STRING, 0)
ZEND_END_ARG_INFO()

ZEND_FUNCTION(register_param_matcher);
