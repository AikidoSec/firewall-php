#include "Includes.h"

void CallPhpExit() {
    AIKIDO_LOG_INFO("Called 'exit'!\n");
#if PHP_VERSION_ID >= 80000
    zend_throw_unwind_exit();
#else
    zend_bailout();
#endif
}

bool CallPhpEcho(std::string message) {
    unsigned int wrote = zend_write(message.c_str(), message.length());  // echo '<message>'
    AIKIDO_LOG_INFO("Called 'echo' -> result %d\n", wrote == message.length());
    return wrote == message.length();
}

bool CallPhpFunction(std::string function_name, unsigned int params_number, zval* params, zval* return_value, zval* object) {
    if (!object && !zend_hash_str_exists(CG(function_table), function_name.c_str(), function_name.size())) {
        AIKIDO_LOG_INFO("Function name '%s' does not exist!\n", function_name.c_str());
        return false;
    }

    zval _function_name;
    zend_string* _function_name_str = zend_string_init(function_name.c_str(), function_name.length(), 0);
    if (!_function_name_str) {
        return false;
    }
    ZVAL_STR(&_function_name, _function_name_str);

    zval* _return_value = return_value;
    zval _temp_return_value;
    if (!return_value) {
        _return_value = &_temp_return_value;
    }

    int _result = call_user_function(EG(function_table), object, &_function_name, _return_value, params_number, params);

    zend_string_release(_function_name_str);

    if (!return_value) {
        zval_ptr_dtor(&_temp_return_value);
    }
    AIKIDO_LOG_INFO("Called user function '%s' -> result %d\n", function_name.c_str(), _result == SUCCESS);
    return _result == SUCCESS;
}

bool CallPhpFunctionWithOneParam(std::string function_name, long first_param, zval* return_value, zval* object) {
    zval _params[1];
    ZVAL_LONG(&_params[0], first_param);
    return CallPhpFunction(function_name, 1, _params, return_value, object);
}

bool CallPhpFunctionWithOneParam(std::string function_name, std::string first_param, zval* return_value, zval* object) {
    zval _params[1];
    zend_string* _first_param = zend_string_init(first_param.c_str(), first_param.length(), 0);
    if (!_first_param) {
        return false;
    }
    ZVAL_STR(&_params[0], _first_param);

    bool ret = CallPhpFunction(function_name, 1, _params, return_value, object);

    zend_string_release(_first_param);

    return ret;
}

std::string CallPhpFunctionCurlGetInfo(zval* curl_handle, int curl_info_option) {
    zval retval;
    zval params[2];

    ZVAL_COPY(&params[0], curl_handle);
    ZVAL_LONG(&params[1], curl_info_option);

    std::string result = "";
    if (CallPhpFunction("curl_getinfo", 2, params, &retval)) {
        switch (Z_TYPE(retval)) {
            case IS_LONG:
                result = std::to_string(Z_LVAL(retval));
                break;
            case IS_STRING:
                result = Z_STRVAL(retval);
                break;
        }
    }

    zval_dtor(&params[0]);
    zval_dtor(&params[1]);
    zval_dtor(&retval);

    return result;
}
