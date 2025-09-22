#include "Includes.h"

#define GET_SERVER_VAR() \
    zval* serverVars = this->GetServerVar(); \
    if (!serverVars) { \
        return ""; \
    }

Server server;

/* Always load the current "_SERVER" variable from PHP, 
so we make sure it's always available and it's the correct one */
zval* Server::GetServerVar() {
    /* Guarantee that "_SERVER" PHP global variable is initialized for the current request */
    if (!zend_is_auto_global_str(ZEND_STRL("_SERVER"))) {
        AIKIDO_LOG_WARN("'_SERVER' autoglobal is not initialized!");
        return nullptr;
    }

    /* Make sure that "_SERVER" PHP global variable is an array */
    if (Z_TYPE(PG(http_globals)[TRACK_VARS_SERVER]) != IS_ARRAY) {
        return nullptr;
    }

    /* Get the "_SERVER" PHP global variable */
    return &PG(http_globals)[TRACK_VARS_SERVER];
}

std::string Server::GetVar(const char* var) {
    GET_SERVER_VAR();
    zval* data = zend_hash_str_find(Z_ARRVAL_P(serverVars), var, strlen(var));
    if (!data) {
        return "";
    }
    return Z_STRVAL_P(data);
}

// Symfony/Component/HttpFoundation/Request class
// Methood getMethod() from this class is used to get the request method
// and also does the method override check, so we need to check if the class is loaded
bool isSymfonyRequestClassLoaded() {
    // check if the class Symfony\Component\HttpFoundation\Request is loaded
    if (zend_lookup_class(
        zend_string_init("Symfony\\Component\\HttpFoundation\\Request", sizeof("Symfony\\Component\\HttpFoundation\\Request")-1, 0)
    ) != NULL) {
        php_printf("Symfony\\Component\\HttpFoundation\\Request class is loaded\n");
        return true;
    }
    php_printf("Symfony\\Component\\HttpFoundation\\Request class is not loaded\n");
    return false;
}

// Return the method from the query param _method (_GET["_method"])
std::string Server::getMethodFromQuery() {
    zval *get_array;
    get_array = zend_hash_str_find(&EG(symbol_table), "_GET", sizeof("_GET") - 1);
    if (!get_array) {
        return "";
    }
    std::string query_method = ToUppercase(Z_STRVAL_P(zend_hash_str_find(Z_ARRVAL_P(get_array), "_method", sizeof("_method") - 1)));
    if (query_method != "") {
        return query_method;
    }
    return "";

}

// For frameworks like Symfony, Laravel, method override is supported using X-HTTP-METHOD-OVERRIDE or _method query param
// https://github.com/symfony/symfony/blob/b8eaa4be31f2159918e79e5694bc9ff241e0d692/src/Symfony/Component/HttpFoundation/Request.php#L1169-L1215
std::string Server::GetMethod() {
    std::string method = ToUppercase(this->GetVar("REQUEST_METHOD"));


    // if (!isSymfonyRequestClassLoaded()) {
    //     return method;
    // }
    
    if (method != "POST") {
        return method;
    }

    // X-HTTP-METHOD-OVERRIDE
    std::string x_http_method_override = ToUppercase(this->GetVar("HTTP_X_HTTP_METHOD_OVERRIDE"));
    if (x_http_method_override != "") {
        method = x_http_method_override;
    }

    // in case of X-HTTP-METHOD-OVERRIDE is not set, we check the query param _method
    if (x_http_method_override == "") {
        std::string query_method = getMethodFromQuery();
        if (query_method != "") {
            method = query_method;
        }
    }
   
    return method;
}

std::string Server::GetRoute() {
    std::string route = this->GetVar("REQUEST_URI");
    size_t pos = route.find("?");
    if (pos != std::string::npos) {
        route = route.substr(0, pos);
    }
    return route;
}

std::string Server::GetStatusCode() {
    return std::to_string(SG(sapi_headers).http_response_code);
}

std::string Server::GetUrl() {
    return (IsHttps() ? "https://" : "http://") + GetVar("HTTP_HOST") + GetVar("REQUEST_URI");
}

std::string Server::GetPost() {
    zval *post_array;
    post_array = zend_hash_str_find(&EG(symbol_table), "_POST", sizeof("_POST") - 1);
    if (!post_array) {
        return "";
    }

    return ArrayToJson(post_array);
}

std::string Server::GetBody() {
    // for application/x-www-form-urlencoded or multipart/form-data, _POST is used
    if(!strncasecmp(GetVar("CONTENT_TYPE").c_str(), "application/x-www-form-urlencoded", strlen("application/x-www-form-urlencoded"))
     || !strncasecmp(GetVar("CONTENT_TYPE").c_str(), "multipart/form-data", strlen("multipart/form-data"))) {
        return this->GetPost();
    }

    long maxlen = PHP_STREAM_COPY_ALL;
    zend_string* contents;
    php_stream* stream;

    stream = php_stream_open_wrapper("php://input", "rb", 0 | REPORT_ERRORS, NULL);
    if ((contents = php_stream_copy_to_mem(stream, maxlen, 0)) != NULL) {
        php_stream_close(stream);
        return std::string(ZSTR_VAL(contents));
    }
    php_stream_close(stream);
    return "";
}

/**
 * Converts the current HTTP query parameters (_GET) into a JSON-formatted string.
 *
 * This function retrieves the query parameters from the `_GET` global array in PHP 
 * and constructs a JSON object representation of the parameters. It supports both 
 * scalar values (e.g., "key=value") and array values (e.g., "key[]=value1&key[]=value2").
 * This function is implemented by just accessing the query that is already parsed by PHP.
 * In this way, we interpret the query in the exact same way that the PHP app receives 
 * the query params.
*/
std::string Server::GetQuery() {
    zval *get_array;
    get_array = zend_hash_str_find(&EG(symbol_table), "_GET", sizeof("_GET") - 1);
    if (!get_array) {
        return "";
    }

    return ArrayToJson(get_array);
}

std::string Server::GetHeaders() {
    GET_SERVER_VAR();
    std::map<std::string, std::string> headers;
    zend_string* key;
    zval* val;
    ZEND_HASH_FOREACH_STR_KEY_VAL(Z_ARRVAL_P(serverVars), key, val) {
        if (key && Z_TYPE_P(val) == IS_STRING) {
            std::string header_name(ZSTR_VAL(key));
            std::string http_header_key;
            std::string http_header_value(Z_STRVAL_P(val));

            if (header_name.find("HTTP_") == 0) {
                http_header_key = header_name.substr(5);
            } else if (header_name == "CONTENT_TYPE" || header_name == "CONTENT_LENGTH" || header_name == "AUTHORIZATION") {
                http_header_key = header_name;
            }

            if (!http_header_key.empty()) {
                std::transform(http_header_key.begin(), http_header_key.end(), http_header_key.begin(), ::tolower);
                headers[http_header_key] = http_header_value;
            }
        }
    }
    ZEND_HASH_FOREACH_END();

    json headers_json;
    for (auto const& [key, val] : headers) {
        headers_json[key] = val;
    }
    return NormalizeAndDumpJson(headers_json);
}

bool Server::IsHttps() {
    return GetVar("HTTPS") != "" ? true : false;
}
