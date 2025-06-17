#include "Includes.h"

std::string GetPhpPackageVersion(const std::string& packageName) {
    zval return_value;
    CallPhpFunctionWithOneParam("phpversion", packageName, &return_value);
    if (Z_TYPE(return_value) == IS_STRING) {
        return Z_STRVAL(return_value);
    }
    return "";
}

unordered_map<std::string, std::string> GetPhpPackages() {
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
            packages[Z_STRVAL_P(extension)] = GetPhpPackageVersion(Z_STRVAL_P(extension));
        }
    } ZEND_HASH_FOREACH_END();

    zval_ptr_dtor(&extensions_array);
    return packages;
}

std::string GetComposerPackageVersion(const std::string& version) {
    if (version.empty()) {
        return version;
    }
    
    if (version[0] == 'v') {
        return version.substr(1);
    }
    
    return version;
}

unordered_map<std::string, std::string> GetComposerPackages() {
    unordered_map<std::string, std::string> packages;

    std::string docRoot = server.GetVar("DOCUMENT_ROOT");
    if (docRoot.empty()) {
        return packages;
    }
    std::string composerLockPath = docRoot + "/../composer.lock";

    std::ifstream composerLockFile(composerLockPath);
    if (!composerLockFile.is_open()) {
        return packages;
    }

    try {
        json composerLockData = json::parse(composerLockFile);
        if (!composerLockData.contains("packages")) {
            return packages;
        }

        const auto& composerLockPackages = composerLockData["packages"];

        for (const auto& composerLockPackage : composerLockPackages) {
            if (!composerLockPackage.contains("name") || !composerLockPackage.contains("version")) {
                continue;
            }
            packages[composerLockPackage["name"]] = GetComposerPackageVersion(composerLockPackage["version"]);
        }
    } catch (const std::exception& e) {
        AIKIDO_LOG_ERROR("Error parsing composer.lock: %s\n", e.what());
        return packages;
    }

    return packages;
}

unordered_map<std::string, std::string> GetPackages() {
    unordered_map<std::string, std::string> packages = GetPhpPackages();
    unordered_map<std::string, std::string> composerPackages = GetComposerPackages();

    packages.insert(composerPackages.begin(), composerPackages.end());

    return packages;
}