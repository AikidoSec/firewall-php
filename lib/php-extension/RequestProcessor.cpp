#include "Includes.h"

RequestProcessor requestProcessor;

std::string RequestProcessor::GetInitData(const std::string& userProvidedToken) {
    LoadLaravelEnvFile();
    LoadEnvironment();

    auto& globalToken = AIKIDO_GLOBAL(token);
    if (!userProvidedToken.empty()) {
        globalToken = userProvidedToken;
    }
    unordered_map<std::string, std::string> packages = GetPackages();
    AIKIDO_GLOBAL(uses_symfony_http_foundation) = packages.find("symfony/http-foundation") != packages.end();
    json initData = {
        {"token", globalToken},
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

bool RequestProcessorInstance::ContextInit() {
    if (!this->requestInitialized || requestProcessor.requestProcessorContextInitFn == nullptr || this->requestProcessorInstance == nullptr) {
        return false;
    }
    return requestProcessor.requestProcessorContextInitFn(this->requestProcessorInstance, GoContextCallback);
}

bool RequestProcessorInstance::SendEvent(EVENT_ID eventId, std::string& output) {
    if (!this->requestInitialized || requestProcessor.requestProcessorOnEventFn == nullptr || this->requestProcessorInstance == nullptr) {
        return false;
    }

    AIKIDO_LOG_DEBUG("Sending event %s...\n", GetEventName(eventId));

    char* charPtr = requestProcessor.requestProcessorOnEventFn(this->requestProcessorInstance, eventId);
    if (!charPtr) {
        AIKIDO_LOG_DEBUG("Got event reply: nullptr\n");
        return true;
    }

    AIKIDO_LOG_DEBUG("Got event reply: %s\n", charPtr);

    output = charPtr;
    free(charPtr);
    return true;
}

void RequestProcessorInstance::SendPreRequestEvent() {
    try {
        std::string outputEvent;
        SendEvent(EVENT_PRE_REQUEST, outputEvent);
        AIKIDO_GLOBAL(action).Execute(outputEvent);
    } catch (const std::exception& e) {
        AIKIDO_LOG_ERROR("Exception encountered in processing request init metadata: %s\n", e.what());
    }
}

void RequestProcessorInstance::SendPostRequestEvent() {
    try {
        std::string outputEvent;
        SendEvent(EVENT_POST_REQUEST, outputEvent);
        AIKIDO_GLOBAL(action).Execute(outputEvent);
    } catch (const std::exception& e) {
        AIKIDO_LOG_ERROR("Exception encountered in processing request shutdown metadata: %s\n", e.what());
    }
}

/*
    If the blocking mode is set from agent (different than -1), return that.
        Otherwise, return the env variable AIKIDO_BLOCK.
*/
bool RequestProcessorInstance::IsBlockingEnabled() {
    if (!this->requestInitialized || requestProcessor.requestProcessorGetBlockingModeFn == nullptr || this->requestProcessorInstance == nullptr) {
        return false;
    }
    int ret = requestProcessor.requestProcessorGetBlockingModeFn(this->requestProcessorInstance);
    if (ret == -1) {
        ret = AIKIDO_GLOBAL(blocking);
    }
    if (ret == 1) {
        AIKIDO_LOG_INFO("Blocking is enabled!\n");
    }
    return ret;
}

bool RequestProcessorInstance::ReportStats() {
    if (requestProcessor.requestProcessorReportStatsFn == nullptr) {
        return false;
    }
    AIKIDO_LOG_INFO("Reporting stats to Aikido Request Processor...\n");

    auto& statsMap = AIKIDO_GLOBAL(stats);
    for (std::unordered_map<std::string, SinkStats>::const_iterator it = statsMap.begin(); it != statsMap.end(); ++it) {
        const std::string& sink = it->first;
        const SinkStats& sinkStats = it->second;
        AIKIDO_LOG_INFO("Reporting stats for sink \"%s\" to Aikido Request Processor...\n", sink.c_str());
        requestProcessor.requestProcessorReportStatsFn(
            this->requestProcessorInstance,
            GoCreateString(sink),
            GoCreateString(sinkStats.kind),
            sinkStats.attacksDetected,
            sinkStats.attacksBlocked,
            sinkStats.interceptorThrewError,
            sinkStats.withoutContext,
            static_cast<GoInt>(sinkStats.timings.size()),
            GoCreateSlice(sinkStats.timings)
        );
    }
    statsMap.clear();
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
    if (AIKIDO_GLOBAL(disable) == true && AIKIDO_GLOBAL(sapi_name) != "apache2handler" && AIKIDO_GLOBAL(sapi_name) != "frankenphp") {
        /* 
            As you can set AIKIDO_DISABLE per site, in an apache-mod-php or frankenphp setup, as a process can serve multiple sites,
            we can't just not initialize the request processor, as it can be disabled for one site but not for another.
            When subsequent requests come in for the non-disabled sites, the request processor needs to be initialized.
            For non-apache-mod-php/frankenphp SAPI, we can just not initialize the request processor if AIKIDO_DISABLE is set to 1.
        */
        AIKIDO_LOG_INFO("Request Processor initialization skipped because AIKIDO_DISABLE is set to 1 and SAPI is not apache2handler or frankenphp!\n");
        return false;
    }

    std::string requestProcessorLibPath = "/opt/aikido-" + std::string(PHP_AIKIDO_VERSION) + "/aikido-request-processor.so";
    this->libHandle = dlopen(requestProcessorLibPath.c_str(), RTLD_LAZY);
    if (!this->libHandle) {
        const char* err = dlerror();
        AIKIDO_LOG_ERROR("Error loading the Aikido Request Processor library from %s: %s!\n", requestProcessorLibPath.c_str(), err);
        this->initFailed = true;
        return false;
    }

    AIKIDO_LOG_INFO("Initializing Aikido Request Processor...\n");

    this->createInstanceFn = (CreateInstanceFn)dlsym(libHandle, "CreateInstance");
    this->destroyInstanceFn = (DestroyInstanceFn)dlsym(libHandle, "DestroyInstance");

    RequestProcessorInitFn requestProcessorInitFn = (RequestProcessorInitFn)dlsym(libHandle, "RequestProcessorInit");
    this->initInstanceFn = (InitInstanceFn)dlsym(libHandle, "InitInstance");
    this->requestProcessorContextInitFn = (RequestProcessorContextInitFn)dlsym(libHandle, "RequestProcessorContextInit");
    this->requestProcessorConfigUpdateFn = (RequestProcessorConfigUpdateFn)dlsym(libHandle, "RequestProcessorConfigUpdate");
    this->requestProcessorOnEventFn = (RequestProcessorOnEventFn)dlsym(libHandle, "RequestProcessorOnEvent");
    this->requestProcessorGetBlockingModeFn = (RequestProcessorGetBlockingModeFn)dlsym(libHandle, "RequestProcessorGetBlockingMode");
    this->requestProcessorReportStatsFn = (RequestProcessorReportStats)dlsym(libHandle, "RequestProcessorReportStats");
    this->requestProcessorUninitFn = (RequestProcessorUninitFn)dlsym(libHandle, "RequestProcessorUninit");
    if (!this->createInstanceFn ||
        !this->initInstanceFn ||
        !this->destroyInstanceFn ||
        !requestProcessorInitFn ||
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
        AIKIDO_LOG_ERROR("Failed to initialize Aikido Request Processor!\n");
        this->initFailed = true;
        return false;
    }

    AIKIDO_GLOBAL(logger).Init();

    AIKIDO_LOG_INFO("Aikido Request Processor initialized successfully (SAPI: %s)!\n", AIKIDO_GLOBAL(sapi_name).c_str());
    return true;
}

bool RequestProcessorInstance::RequestInit() {
    std::string sapiName = sapi_module.name;
    if (sapiName == "frankenphp") {
        if (GetEnvBool("FRANKENPHP_WORKER", false)) {
            AIKIDO_GLOBAL(isWorkerMode) = true;
            AIKIDO_LOG_INFO("FrankenPHP worker warm-up request detected, skipping RequestInit\n");
            return true;
        }
    }
    
    if (this->requestProcessorInstance == nullptr && requestProcessor.createInstanceFn != nullptr) {
        this->threadId = GetThreadID();

        this->requestProcessorInstance = requestProcessor.createInstanceFn(this->threadId);
        if (this->requestProcessorInstance == nullptr) {
            AIKIDO_LOG_ERROR("Failed to create Go RequestProcessorInstance!\n");
            return false;
        }
        if(!requestProcessor.initInstanceFn(this->requestProcessorInstance, GoCreateString(requestProcessor.GetInitData()))) {
            AIKIDO_LOG_ERROR("Failed to initialize Go RequestProcessorInstance!\n");
            return false;
        }

        AIKIDO_LOG_INFO("Created Go RequestProcessorInstance (threadId: %lu)\n", this->threadId);
    }
    
    if (sapiName == "apache2handler" || sapiName == "frankenphp") {
      // Apache-mod-php and FrankenPHP can serve multiple sites per process
      // We need to reload config each request to detect token changes
        this->LoadConfigFromEnvironment();
    } else {
        // Server APIs that are not apache-mod-php/frankenphp (like php-fpm, cli-server, ...) 
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
        AIKIDO_GLOBAL(requestProcessorInstance).ReportStats();
    }
    return true;
}

void RequestProcessorInstance::LoadConfig(const std::string& previousToken, const std::string& currentToken) {
    if (requestProcessor.requestProcessorConfigUpdateFn == nullptr || this->requestProcessorInstance == nullptr) {
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
    std::string initJson = requestProcessor.GetInitData(currentToken);
    requestProcessor.requestProcessorConfigUpdateFn(this->requestProcessorInstance, GoCreateString(initJson));
}

void RequestProcessorInstance::LoadConfigFromEnvironment() {
    auto& globalToken = AIKIDO_GLOBAL(token);
    std::string previousToken = globalToken;
        
    LoadEnvironment();
    
    std::string currentToken = globalToken;
    
    LoadConfig(previousToken, currentToken);
}

void RequestProcessorInstance::LoadConfigWithTokenFromPHPSetToken(const std::string& tokenFromMiddleware) {
    LoadConfig(AIKIDO_GLOBAL(token), tokenFromMiddleware);
}

void RequestProcessorInstance::RequestShutdown() {
    SendPostRequestEvent();
    this->requestInitialized = false;
}

void RequestProcessor::Uninit() {
    if (!requestProcessor.libHandle) {
        return;
    }
    dlclose(this->libHandle);
    this->libHandle = nullptr;
    AIKIDO_LOG_INFO("Aikido Request Processor unloaded!\n");
}

RequestProcessorInstance::~RequestProcessorInstance() {
    if (!requestProcessor.initFailed && requestProcessor.requestProcessorUninitFn && this->requestProcessorInstance != nullptr) {
        AIKIDO_LOG_INFO("Reporting final stats to Aikido Request Processor...\n");
        this->ReportStats();

        AIKIDO_LOG_INFO("Calling uninit for Aikido Request Processor...\n");
        requestProcessor.requestProcessorUninitFn(this->requestProcessorInstance);
    }
    if (requestProcessor.destroyInstanceFn && this->requestProcessorInstance != nullptr) {
        AIKIDO_LOG_INFO("Destroying Go RequestProcessorInstance (threadId: %lu)...\n", this->threadId);
        requestProcessor.destroyInstanceFn(this->threadId);
        this->requestProcessorInstance = nullptr;
    }
}