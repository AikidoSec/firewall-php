#include "Includes.h"

RequestProcessor requestProcessor;

std::string RequestProcessor::GetInitData(std::string token) {
    LoadLaravelEnvFile();
    LoadEnvironment();

    if (token.empty()) {
        token = AIKIDO_GLOBAL(token);
    }

    json initData = {
        {"socket_path", AIKIDO_GLOBAL(socket_path)},
        {"token", token},
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
        {"packages", GetPackages()}};
    return NormalizeAndDumpJson(initData);
}

bool RequestProcessor::ContextInit() {
    return this->requestProcessorContextInitFn(GoContextCallback);
}

bool RequestProcessor::SendEvent(EVENT_ID eventId, std::string& output) {
    if (!this->requestInitialized) {
        return false;
    }

    AIKIDO_LOG_DEBUG("Sending event %s...\n", GetEventName(eventId));

    char* charPtr = requestProcessorOnEventFn(eventId);
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
    if (!this->requestInitialized) {
        return false;
    }
    int ret = requestProcessorGetBlockingModeFn();
    if (ret == -1) {
        return AIKIDO_GLOBAL(blocking);
    }
    return ret;
}

bool RequestProcessor::ReportStats() {
    AIKIDO_LOG_INFO("Reporting stats to Aikido Request Processor...\n");

    for (const auto& [sink, sinkStats] : stats) {
        AIKIDO_LOG_INFO("Reporting stats for sink \"%s\" to Aikido Request Processor...\n", sink.c_str());
        requestProcessorReportStatsFn(GoCreateString(sink), GoCreateString(sinkStats.kind), sinkStats.attacksDetected, sinkStats.attacksBlocked, sinkStats.interceptorThrewError, sinkStats.withoutContext, sinkStats.timings.size(), GoCreateSlice(sinkStats.timings));
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

    if (AIKIDO_GLOBAL(disable) == true) {
        AIKIDO_LOG_INFO("Request Processor initialization skipped because AIKIDO_DISABLE is set to 1!\n");
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

    std::string initDataString = this->GetInitData();
    if (!requestProcessorInitFn(GoCreateString(initDataString))) {
        AIKIDO_LOG_ERROR("Failed to initialize Aikido Request Processor library: %s!\n", dlerror());
        this->initFailed = true;
        return false;
    }

    AIKIDO_GLOBAL(logger).Init();

    AIKIDO_LOG_INFO("Aikido Request Processor initialized successfully!\n");
    return true;
}

bool RequestProcessor::RequestInit() {
    if (!this->Init()) {
        AIKIDO_LOG_ERROR("Failed to initialize the request processor: %s!\n", dlerror());
        return false;
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

void RequestProcessor::LoadConfig(std::string token) {
    if (token.empty() && this->configReloaded) {
        return;
    }
    
    AIKIDO_LOG_INFO("Reloading Aikido config...\n");
    std::string initJson = this->GetInitData(token);
    this->requestProcessorConfigUpdateFn(GoCreateString(initJson));
    this->configReloaded = true;
}

void RequestProcessor::RequestShutdown() {
    LoadConfig();
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