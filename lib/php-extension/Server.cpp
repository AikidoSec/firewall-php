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
    if (!zend_is_auto_global_str(ZEND_STRL("_SERVER"));) {
        AIKIDO_LOG_WARN("'_SERVER' autoglobal is not initialized!");
        return nullptr;
    }

    /* Make sure that "_SERVER" PHP global variable is an array */
    if (Z_TYPE(PG(http_globals)[TRACK_VARS_SERVER]) != IS_ARRAY) {
        return nullptr;
    }

    /* Get the "_SERVER" PHP global variable */
    return PG(http_globals)[TRACK_VARS_SERVER];
}

std::string Server::GetVar(const char* var) {
    GET_SERVER_VAR();
    zval* data = zend_hash_str_find(Z_ARRVAL_P(serverVars), var, strlen(var));
    if (!data) {
        return "";
    }
    return Z_STRVAL_P(data);
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

std::string Server::GetBody() {
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

    json query_json;
    zend_string *key;
    zval *val;
    ZEND_HASH_FOREACH_STR_KEY_VAL(Z_ARRVAL_P(get_array), key, val) {
        if(key && val) {
            std::string key_str(ZSTR_VAL(key));
            if (Z_TYPE_P(val) == IS_STRING) {
                query_json[key_str] = Z_STRVAL_P(val);
            }
            else if (Z_TYPE_P(val) == IS_ARRAY){
                json val_array = json::array();
                zval *v;
                ZEND_HASH_FOREACH_VAL(Z_ARRVAL_P(val), v) {
                    if (Z_TYPE_P(v) == IS_STRING) {
                        val_array.push_back(Z_STRVAL_P(v));
                    }
                } 
                ZEND_HASH_FOREACH_END();
                query_json[key_str] = val_array;
            }
        }
    }
    ZEND_HASH_FOREACH_END();

    return NormalizeAndDumpJson(query_json);
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
