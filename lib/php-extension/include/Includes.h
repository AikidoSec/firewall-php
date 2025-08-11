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
#include <spawn.h>
#include <fstream>
#include <iostream>

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
#include "Stats.h"
#include "Hooks.h"
#include "PhpWrappers.h"
#include "Server.h"
#include "RequestProcessor.h"
#include "PhpLifecycle.h"
#include "Packages.h"

#include "Utils.h"

#include "Handle.h"
#include "HandleUsers.h"
#include "HandleSetToken.h"
#include "HandleUrls.h"
#include "HandleShellExecution.h"
#include "HandleShouldBlockRequest.h"
#include "HandleSetRateLimitGroup.h"
#include "HandleQueries.h"
#include "HandlePathAccess.h"
#include "HandleFileCompilation.h"
#include "HookAst.h"