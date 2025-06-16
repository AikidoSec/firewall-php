#include "Includes.h"

std::string GetPackageVersion(const std::string& packageName) {
    zval return_value;
    CallPhpFunctionWithOneParam("phpversion", packageName, &return_value);
    if (Z_TYPE(return_value) == IS_STRING) {
        return Z_STRVAL(return_value);
    }
    return "";
}

unordered_map<std::string, std::string> GetPackages() {
    unordered_map<std::string, std::string> packages;

    zval extensions_array;
    if (!CallPhpFunction("get_loaded_extensions", 0, nullptr, &extensions_array)) {
        return packages;
    }

    if (Z_TYPE(extensions_array) != IS_ARRAY) {
        return packages;
    }

    zend_string *key;
    zval *extension;
    zend_ulong index;

    ZEND_HASH_FOREACH_KEY_VAL(Z_ARRVAL(extensions_array), index, key, extension) {
        if (extension && Z_TYPE_P(extension) == IS_STRING) {
            packages[Z_STRVAL_P(extension)] = GetPackageVersion(Z_STRVAL_P(extension));
            AIKIDO_LOG_INFO("Found package %s version %s\n", Z_STRVAL_P(extension), packages[Z_STRVAL_P(extension)].c_str());
        }
    } ZEND_HASH_FOREACH_END();

    zval_ptr_dtor(&extensions_array);
    return packages;
}