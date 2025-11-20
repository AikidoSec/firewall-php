#include "Includes.h"

RequestProcessor requestProcessor;

std::string RequestProcessor::GetInitData(const std::string& token) {
    LoadLaravelEnvFile();
    LoadEnvironment();

    if (!token.empty()) {
        AIKIDO_GLOBAL(token) = token;
    }
    unordered_map<std::string, std::string> packages = GetPackages();
    AIKIDO_GLOBAL(uses_symfony_http_foundation) = packages.find("symfony/http-foundation") != packages.end();
    json initData = {
        {"token", AIKIDO_GLOBAL(token)},
        {"platform_name", AIKIDO_GLOBAL(sapi_name)},
        {"platform_version", PHP_VERSION},
        {"endpoint", AIKIDO_GLOBAL(endpoint)},
        {"config_endpoint", AIKIDO_GLOBAL(config_endpoint)},
        {"log_level", AIKIDO_GLOBAL(log_level_str)},
        {"blocking", AIKIDO_GLOBAL(blocking)},
        {"trust_proxy", AIKIDO_GLOBAL(trust_proxy)},
        {"disk_logs", AIKIDO_GLOBAL(disk_logs)},
        {"localhost_allowed_by_default", AIKIDO_GLOBAL(localhost_allowed_by_default)},
        {"collect_api_schema", AIKIDO_GLOBAL(collect_api_schema)},
        {"packages", packages}};
    return NormalizeAndDumpJson(initData);
}

bool RequestProcessor::ContextInit() {
    if (!this->requestInitialized || this->requestProcessorContextInitFn == nullptr) {
        return false;
    }
    return this->requestProcessorContextInitFn(GoContextCallback);
}

bool RequestProcessor::SendEvent(EVENT_ID eventId, std::string& output) {
    if (!this->requestInitialized || this->requestProcessorOnEventFn == nullptr) {
        return false;
    }

    AIKIDO_LOG_DEBUG("Sending event %s...\n", GetEventName(eventId));

    char* charPtr = this->requestProcessorOnEventFn(eventId);
    if (!charPtr) {
        AIKIDO_LOG_DEBUG("Got event reply: nullptr\n");
        return true;
    }

    AIKIDO_LOG_DEBUG("Got event reply: %s\n", charPtr);

    output = charPtr;
    free(charPtr);
    return true;
}

void RequestProcessor::SendPreRequestEvent() {
    try {
        std::string outputEvent;
        SendEvent(EVENT_PRE_REQUEST, outputEvent);
        action.Execute(outputEvent);
    } catch (const std::exception& e) {
        AIKIDO_LOG_ERROR("Exception encountered in processing request init metadata: %s\n", e.what());
    }
}

void RequestProcessor::SendPostRequestEvent() {
    try {
        std::string outputEvent;
        SendEvent(EVENT_POST_REQUEST, outputEvent);
        action.Execute(outputEvent);
    } catch (const std::exception& e) {
        AIKIDO_LOG_ERROR("Exception encountered in processing request shutdown metadata: %s\n", e.what());
    }
}

/*
    If the blocking mode is set from agent (different than -1), return that.
        Otherwise, return the env variable AIKIDO_BLOCK.
*/
bool RequestProcessor::IsBlockingEnabled() {
    if (!this->requestInitialized || this->requestProcessorGetBlockingModeFn == nullptr) {
        return false;
    }
    int ret = this->requestProcessorGetBlockingModeFn();
    if (ret == -1) {
        ret = AIKIDO_GLOBAL(blocking);
    }
    if (ret == 1) {
        AIKIDO_LOG_INFO("Blocking is enabled!\n");
    }
    return ret;
}

bool RequestProcessor::ReportStats() {
    if (this->requestProcessorReportStatsFn == nullptr) {
        return false;
    }
    AIKIDO_LOG_INFO("Reporting stats to Aikido Request Processor...\n");

    for (const auto& [sink, sinkStats] : stats) {
        AIKIDO_LOG_INFO("Reporting stats for sink \"%s\" to Aikido Request Processor...\n", sink.c_str());
        this->requestProcessorReportStatsFn(GoCreateString(sink), GoCreateString(sinkStats.kind), sinkStats.attacksDetected, sinkStats.attacksBlocked, sinkStats.interceptorThrewError, sinkStats.withoutContext, sinkStats.timings.size(), GoCreateSlice(sinkStats.timings));
    }
    stats.clear();
    return true;
}

bool RequestProcessor::Init() {
    if (this->initFailed) {
        return false;
    }

    if (this->libHandle) {
        return true;
    }

    std::string initDataString = this->GetInitData();
    if (AIKIDO_GLOBAL(disable) == true && AIKIDO_GLOBAL(sapi_name) != "apache2handler") {
        /* 
            As you can set AIKIDO_DISABLE per site, in an apache-mod-php setup, as a process can serve multiple sites,
            we can't just not initialize the request processor, as it can be disabled for one site but not for another.
            When subsequent requests come in for the non-disabled sites, the request processor needs to be initialized.
            For non-apache-mod-php SAPI, we can just not initialize the request processor if AIKIDO_DISABLE is set to 1.
        */
        AIKIDO_LOG_INFO("Request Processor initialization skipped because AIKIDO_DISABLE is set to 1 and SAPI is not apache2handler!\n");
        return false;
    }

    std::string requestProcessorLibPath = "/opt/aikido-" + std::string(PHP_AIKIDO_VERSION) + "/aikido-request-processor.so";
    this->libHandle = dlopen(requestProcessorLibPath.c_str(), RTLD_LAZY);
    if (!this->libHandle) {
        AIKIDO_LOG_ERROR("Error loading the Aikido Request Processor library from %s: %s!\n", requestProcessorLibPath.c_str(), dlerror());
        this->initFailed = true;
        return false;
    }

    AIKIDO_LOG_INFO("Initializing Aikido Request Processor...\n");

    RequestProcessorInitFn requestProcessorInitFn = (RequestProcessorInitFn)dlsym(libHandle, "RequestProcessorInit");
    this->requestProcessorContextInitFn = (RequestProcessorContextInitFn)dlsym(libHandle, "RequestProcessorContextInit");
    this->requestProcessorConfigUpdateFn = (RequestProcessorConfigUpdateFn)dlsym(libHandle, "RequestProcessorConfigUpdate");
    this->requestProcessorOnEventFn = (RequestProcessorOnEventFn)dlsym(libHandle, "RequestProcessorOnEvent");
    this->requestProcessorGetBlockingModeFn = (RequestProcessorGetBlockingModeFn)dlsym(libHandle, "RequestProcessorGetBlockingMode");
    this->requestProcessorReportStatsFn = (RequestProcessorReportStats)dlsym(libHandle, "RequestProcessorReportStats");
    this->requestProcessorUninitFn = (RequestProcessorUninitFn)dlsym(libHandle, "RequestProcessorUninit");
    if (!requestProcessorInitFn ||
        !this->requestProcessorContextInitFn ||
        !this->requestProcessorConfigUpdateFn ||
        !this->requestProcessorOnEventFn ||
        !this->requestProcessorGetBlockingModeFn ||
        !this->requestProcessorReportStatsFn ||
        !this->requestProcessorUninitFn) {
        AIKIDO_LOG_ERROR("Error loading symbols from the Aikido Request Processor library!\n");
        this->initFailed = true;
        return false;
    }

    if (!requestProcessorInitFn(GoCreateString(initDataString))) {
        AIKIDO_LOG_ERROR("Failed to initialize Aikido Request Processor library: %s!\n", dlerror());
        this->initFailed = true;
        return false;
    }

    AIKIDO_GLOBAL(logger).Init();

    AIKIDO_LOG_INFO("Aikido Request Processor initialized successfully (SAPI: %s)!\n", AIKIDO_GLOBAL(sapi_name).c_str());
    return true;
}

bool RequestProcessor::RequestInit() {
    if (!this->Init()) {
        AIKIDO_LOG_ERROR("Failed to initialize the request processor: %s!\n", dlerror());
        return false;
    }
    
    if (AIKIDO_GLOBAL(sapi_name) == "apache2handler") {
      // Apache-mod-php can serve multiple sites per process
      // We need to reload config each request to detect token changes
        this->LoadConfigFromEnvironment();
    } else {
        // Server APIs that are not apache-mod-php (like php-fpm, cli-server, ...) 
        //  can only serve one site per process, so the config should be loaded at the first request.
        // If the token is not set at the first request, we try to reload it until we get a valid token.
        // The user can update .env file via zero downtime deployments after the PHP server is started.
        if (AIKIDO_GLOBAL(token) == "") {
            AIKIDO_LOG_INFO("Loading Aikido config until we get a valid token for SAPI: %s...\n", AIKIDO_GLOBAL(sapi_name).c_str());
            this->LoadConfigFromEnvironment();
        }
    }

    AIKIDO_LOG_DEBUG("RINIT started!\n");

    if (AIKIDO_GLOBAL(disable) == true) {
        AIKIDO_LOG_INFO("Request Processor initialization skipped because AIKIDO_DISABLE is set to 1!\n");
        return true;
    }

    this->requestInitialized = true;
    this->numberOfRequests++;

    ContextInit();
    SendPreRequestEvent();

    if ((this->numberOfRequests % AIKIDO_GLOBAL(report_stats_interval_to_agent)) == 0) {
        requestProcessor.ReportStats();
    }
    return true;
}

void RequestProcessor::LoadConfig(const std::string& previousToken, const std::string& currentToken) {
    if (this->requestProcessorConfigUpdateFn == nullptr) {
        return;
    }
    if (currentToken.empty()) {
        AIKIDO_LOG_INFO("Current token is empty, skipping config reload...!\n");
        return;
    }
    if (previousToken == currentToken) {
        AIKIDO_LOG_INFO("Token is the same as previous one, skipping config reload...\n");
        return;
    }

    AIKIDO_LOG_INFO("Reloading Aikido config...\n");
    std::string initJson = this->GetInitData(currentToken);
    this->requestProcessorConfigUpdateFn(GoCreateString(initJson));
}

void RequestProcessor::LoadConfigFromEnvironment() {
    std::string previousToken = AIKIDO_GLOBAL(token);
    LoadEnvironment();
    std::string currentToken = AIKIDO_GLOBAL(token);
    LoadConfig(previousToken, currentToken);
}

void RequestProcessor::LoadConfigWithTokenFromPHPSetToken(const std::string& tokenFromMiddleware) {
    LoadConfig(AIKIDO_GLOBAL(token), tokenFromMiddleware);
}

void RequestProcessor::RequestShutdown() {
    SendPostRequestEvent();
    this->requestInitialized = false;
}

void RequestProcessor::Uninit() {
    if (!this->libHandle) {
        return;
    }
    if (!this->initFailed && this->requestProcessorUninitFn) {
        AIKIDO_LOG_INFO("Reporting final stats to Aikido Request Processor...\n");
        this->ReportStats();

        AIKIDO_LOG_INFO("Calling uninit for Aikido Request Processor...\n");
        this->requestProcessorUninitFn();
    }
    dlclose(this->libHandle);
    this->libHandle = nullptr;
    AIKIDO_LOG_INFO("Aikido Request Processor unloaded!\n");
}

RequestProcessor::~RequestProcessor() {
    this->Uninit();
}