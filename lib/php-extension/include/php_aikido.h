#pragma once

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

#include <unordered_map>
#include <chrono>
#include "php.h"
#include "Log.h"
#include "Agent.h"
#include "Server.h"
#include "RequestProcessor.h"
#include "Action.h"
#include "Cache.h"
#include "PhpLifecycle.h"
#include "Stats.h"

extern zend_module_entry aikido_module_entry;
#define phpext_aikido_ptr &aikido_module_entry

#define PHP_AIKIDO_VERSION "1.4.5"

#if defined(ZTS) && defined(COMPILE_DL_AIKIDO)
ZEND_TSRMLS_CACHE_EXTERN()
#endif

ZEND_BEGIN_MODULE_GLOBALS(aikido)
bool environment_loaded;
long log_level;
bool blocking;
bool disable;
bool disk_logs; // When enabled, it writes logs to disk instead of stdout. It's usefull when debugging to have the logs grouped by PHP process.
bool collect_api_schema;
bool trust_proxy;
bool localhost_allowed_by_default;
unsigned int report_stats_interval_to_agent; // Report once every X requests the collected stats to Agent
std::string log_level_str;
std::string sapi_name;
std::string token;
std::string endpoint;
std::string config_endpoint;
Log logger;
Agent agent;
Server server;
RequestProcessor requestProcessor;
Action action;
RequestCache requestCache;
EventCache eventCache;
PhpLifecycle phpLifecycle;
std::unordered_map<std::string, SinkStats> stats;
std::chrono::high_resolution_clock::time_point currentRequestStart;
uint64_t totalOverheadForCurrentRequest;
std::unordered_map<std::string, std::string> laravelEnv;
bool laravelEnvLoaded;
bool checkedAutoBlock;
bool checkedShouldBlockRequest;
HashTable *global_ast_to_clean;
void (*original_ast_process)(zend_ast *ast);
ZEND_END_MODULE_GLOBALS(aikido)

ZEND_EXTERN_MODULE_GLOBALS(aikido)

#define AIKIDO_GLOBAL(v) ZEND_MODULE_GLOBALS_ACCESSOR(aikido, v)

/* For compatibility with older PHP versions */
#ifndef ZEND_PARSE_PARAMETERS_NONE
#define ZEND_PARSE_PARAMETERS_NONE()  \
    ZEND_PARSE_PARAMETERS_START(0, 0) \
    ZEND_PARSE_PARAMETERS_END()
#endif
