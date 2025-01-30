#pragma once

extern zend_module_entry aikido_module_entry;
#define phpext_aikido_ptr &aikido_module_entry

#define PHP_AIKIDO_VERSION "1.0.108"

#if defined(ZTS) && defined(COMPILE_DL_AIKIDO)
ZEND_TSRMLS_CACHE_EXTERN()
#endif

ZEND_BEGIN_MODULE_GLOBALS(aikido)
bool environment_loaded;
long log_level;
bool blocking;
bool disable;
bool collect_api_schema;
bool trust_proxy;
bool localhost_allowed_by_default;
unsigned int report_stats_interval_to_agent; // Report once every X requests the collected stats to Agent
std::string socket_path;
std::string log_level_str;
std::string sapi_name;
std::string token;
std::string endpoint;
std::string config_endpoint;
Log logger;
Agent agent;
ZEND_END_MODULE_GLOBALS(aikido)

ZEND_EXTERN_MODULE_GLOBALS(aikido)

#define AIKIDO_GLOBAL(v) ZEND_MODULE_GLOBALS_ACCESSOR(aikido, v)

/* For compatibility with older PHP versions */
#ifndef ZEND_PARSE_PARAMETERS_NONE
#define ZEND_PARSE_PARAMETERS_NONE()  \
    ZEND_PARSE_PARAMETERS_START(0, 0) \
    ZEND_PARSE_PARAMETERS_END()
#endif