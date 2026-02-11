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

#define PHP_AIKIDO_VERSION "1.5.0"

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
bool uses_symfony_http_foundation; // If true, method override is supported using X-HTTP-METHOD-OVERRIDE or _method query param
unsigned int report_stats_interval_to_agent; // Report once every X requests the collected stats to Agent
std::chrono::high_resolution_clock::time_point currentRequestStart;
uint64_t totalOverheadForCurrentRequest;
bool laravelEnvLoaded;
bool checkedAutoBlock;
bool checkedShouldBlockRequest;
// This variable is used to check if the request is bypassed,
// if true, all blocking checks will be skipped.
bool isIpBypassed;
bool isWorkerMode;
HashTable *globalAstToClean;
void (*original_ast_process)(zend_ast *ast);
// IMPORTANT: The order of these objects MUST NOT be changed due to object interdependencies.
// This ensures proper construction/destruction order in both ZTS and non-ZTS modes.
// Objects are constructed in this order and destroyed in reverse order.
std::string log_level_str;
std::string sapi_name;
std::string token;
std::string endpoint;
std::string config_endpoint;
/*
    Cache objects used by the PHP extension to share state with the Go request processor.

    - RequestCache stores data that is scoped to a single incoming HTTP request
      (method, route, user id, rate-limit group, ...). The Go side calls into the
      extension once per request to populate and later clear this structure
      (see context.Init / context.Clear in the Go code).

    - EventCacheStack stores data that is scoped to a single *event*, where an event
      is one hooked PHP function call (e.g. curl_exec, file_get_contents, exec,
      PDO::query, ...). Each hook invocation pushes a new EventCache on the stack
      when it starts and pops it when it finishes, so nested hooks have independent
      contexts.
*/
RequestCache requestCache;
EventCacheStack eventCacheStack;

/*
    Reset helpers:

    These functions re-initialize the cache structs to their default state instead
    of reallocating them. The PHP extension code runs inside long-lived PHP/Apache/FPM
    processes that handle many HTTP requests. Because these cache objects live for
    the lifetime of the process, we must explicitly reset them so that no state
    from one request or event can leak into the next.
*/
EventCache eventCache;
Agent agent;
Log logger;
Server server;
std::unordered_map<std::string, SinkStats> stats;
RequestProcessor requestProcessor;
Action action;
PhpLifecycle phpLifecycle;
std::unordered_map<std::string, std::string> laravelEnv;
ZEND_END_MODULE_GLOBALS(aikido)

ZEND_EXTERN_MODULE_GLOBALS(aikido)

#define AIKIDO_GLOBAL(v) ZEND_MODULE_GLOBALS_ACCESSOR(aikido, v)

/* For compatibility with older PHP versions */
#ifndef ZEND_PARSE_PARAMETERS_NONE
#define ZEND_PARSE_PARAMETERS_NONE()  \
    ZEND_PARSE_PARAMETERS_START(0, 0) \
    ZEND_PARSE_PARAMETERS_END()
#endif
