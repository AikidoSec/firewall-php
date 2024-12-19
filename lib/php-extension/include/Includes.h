#pragma once

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

#include <arpa/inet.h>
#include <curl/curl.h>
#include <ifaddrs.h>
#include <net/if.h>
#include <netinet/in.h>
#include <sys/types.h>
#include <sys/utsname.h>

#include <functional>
#include <random>
#include <string>
#include <ctime>
#include <unordered_map>
#include <chrono>

#include "3rdparty/json.hpp"
using namespace std;
using json = nlohmann::json;

#include "SAPI.h"
#include "ext/pdo/php_pdo_driver.h"
#include "ext/standard/info.h"
#include "php.h"
#include "zend_exceptions.h"

#include "GoCGO.h"
#include "GoWrappers.h"

#include "../../API.h"
#include "Log.h"
#include "Agent.h"
#include "php_aikido.h"
#include "Environment.h"
#include "Action.h"
#include "Cache.h"
#include "Hooks.h"
#include "PhpWrappers.h"
#include "Request.h"
#include "RequestProcessor.h"
#include "PhpLifecycle.h"
#include "Stats.h"

#include "Utils.h"

#include "Handle.h"
#include "HandleUsers.h"
#include "HandleUrls.h"
#include "HandleShellExecution.h"
#include "HandleShouldBlockRequest.h"
#include "HandlePDO.h"
#include "HandlePathAccess.h"
