#include "HandlePathTraversal.h"
#include "Utils.h"

AIKIDO_HANDLER_FUNCTION(handle_file_path_access) {
    zend_string *filename = NULL;
   

    ZEND_PARSE_PARAMETERS_START(0, -1)
        Z_PARAM_OPTIONAL 
        Z_PARAM_STR(filename)
    ZEND_PARSE_PARAMETERS_END();
    if (filename == NULL) {
        return;
    }

    std::string filenameString(ZSTR_VAL(filename));

    std::string functionNameString(AIKIDO_GET_FUNCTION_NAME());
    
    inputEvent = {
        { "event", "function_executed" },
        { "data", {
            { "function_name", "path_accessed" },
            { "parameters", {
                { "filename", filenameString },
                { "filename2", "" },
                { "operation", functionNameString },
                { "context", get_context()}
            } }
        } }
    };
}

AIKIDO_HANDLER_FUNCTION(handle_file_path_access_2) {
    zend_string *filename = NULL;
    zend_string *filename2 = NULL;
   

    ZEND_PARSE_PARAMETERS_START(0, -1)
        Z_PARAM_OPTIONAL 
        Z_PARAM_STR(filename)
        Z_PARAM_STR(filename2)
    ZEND_PARSE_PARAMETERS_END();
    if (filename == NULL || filename2 == NULL) {
        return;
    }

    std::string filenameString(ZSTR_VAL(filename));
    std::string filenameString2(ZSTR_VAL(filename2));

    std::string functionNameString(AIKIDO_GET_FUNCTION_NAME());
    
    inputEvent = {
        { "event", "function_executed" },
        { "data", {
            { "function_name", "path_accessed" },
            { "parameters", {
                { "filename", filenameString },
                { "filename2", filenameString2 },
                { "operation", functionNameString },
                { "context", get_context()}
            } }
        } }
    };   
}

