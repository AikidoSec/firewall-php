#include "Includes.h"

ZEND_FUNCTION(set_token) {
    ScopedTimer scopedTimer("set_token", "aikido_op");

    if (AIKIDO_GLOBAL(disable) == true) {
        RETURN_BOOL(false);
    }

    char* token = nullptr;
    size_t tokenLength = 0;

    ZEND_PARSE_PARAMETERS_START(1, 1)
        Z_PARAM_STRING(token, tokenLength)
    ZEND_PARSE_PARAMETERS_END();

    if (!token || tokenLength == 0) {
        AIKIDO_LOG_ERROR("set_token: token is null!\n");
        RETURN_BOOL(false);
    }

    AIKIDO_GLOBAL(requestProcessor).LoadConfigWithTokenFromPHPSetToken(std::string(token, tokenLength));
    RETURN_BOOL(true);
}
