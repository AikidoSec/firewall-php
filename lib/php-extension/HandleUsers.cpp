#include "Includes.h"

bool SendUserEvent(std::string id, std::string username) {
    requestCache.userId = id;
    requestCache.userName = username;

    try {
        std::string output;
        requestProcessor.SendEvent(EVENT_SET_USER, output);
        action.Execute(output);
        return true;
    } catch (const std::exception &e) {
        AIKIDO_LOG_ERROR("Exception encountered in processing user event: %s\n", e.what());
    }
    return false;
}

// Exports the "\aikido\set_user" function, to be called from PHP user code.
// Receives two parameters: "id" (string) and "name" (string, optional).
// Returns true if the setting of the user succeeded, false otherwise.
ZEND_FUNCTION(set_user) {
    ScopedTimer scopedTimer("set_user");

    if (AIKIDO_GLOBAL(disable) == true) {
        RETURN_BOOL(false);
    }

    requestProcessor.LoadConfigOnce();

    char* id = nullptr;
    size_t idLength = 0;
    char* name = nullptr;
    size_t nameLength = 0;

    ZEND_PARSE_PARAMETERS_START(1, 2)
        Z_PARAM_STRING(id, idLength)
        Z_PARAM_OPTIONAL
        Z_PARAM_STRING(name, nameLength)
    ZEND_PARSE_PARAMETERS_END();

    std::string idString = std::string(id, idLength);
    std::string nameString = "";
    if (name && nameLength) {
        nameString = std::string(name, nameLength);
    }

    RETURN_BOOL(SendUserEvent(idString, nameString));
}
