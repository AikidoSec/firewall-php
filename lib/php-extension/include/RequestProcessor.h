#pragma once

typedef void* (*CreateInstanceFn)(uint64_t threadId);
typedef GoUint8 (*InitInstanceFn)(void* instancePtr, GoString initJson);
typedef void (*DestroyInstanceFn)(uint64_t threadId);

// Updated typedefs with instance pointer as first parameter
typedef GoUint8 (*RequestProcessorInitFn)(GoString platformName);
typedef GoUint8 (*RequestProcessorContextInitFn)(void* instancePtr, ContextCallback);
typedef GoUint8 (*RequestProcessorConfigUpdateFn)(void* instancePtr, GoString initJson);
typedef char* (*RequestProcessorOnEventFn)(void* instancePtr, GoInt eventId);
typedef int (*RequestProcessorGetBlockingModeFn)(void* instancePtr);
typedef void (*RequestProcessorReportStats)(void* instancePtr, GoString, GoString, GoInt32, GoInt32, GoInt32, GoInt32, GoInt32, GoSlice);
typedef void (*RequestProcessorUninitFn)(void* instancePtr);

class RequestProcessor {
    #ifdef ZTS
        private:
            std::mutex syncMutex;
    #endif
    
    public:
    bool initFailed = false;
    void* libHandle = nullptr;

    CreateInstanceFn createInstanceFn = nullptr;
    InitInstanceFn initInstanceFn = nullptr;
    DestroyInstanceFn destroyInstanceFn = nullptr;
    RequestProcessorContextInitFn requestProcessorContextInitFn = nullptr;
    RequestProcessorConfigUpdateFn requestProcessorConfigUpdateFn = nullptr;
    RequestProcessorOnEventFn requestProcessorOnEventFn = nullptr;
    RequestProcessorGetBlockingModeFn requestProcessorGetBlockingModeFn = nullptr;
    RequestProcessorReportStats requestProcessorReportStatsFn = nullptr;
    RequestProcessorUninitFn requestProcessorUninitFn = nullptr;

    RequestProcessor() = default;
    ~RequestProcessor();

    std::string GetInitData(const std::string& userProvidedToken = "");

    bool Init();
    void Uninit();
};

class RequestProcessorInstance {
   private:
    bool requestInitialized = false;
    void* requestProcessorInstance = nullptr; 
    uint64_t numberOfRequests = 0;
    uint64_t threadId = 0;

    bool ContextInit();
    void SendPreRequestEvent();
    void SendPostRequestEvent();

   public:
    RequestProcessorInstance() = default;

    bool RequestInit();
    bool SendEvent(EVENT_ID eventId, std::string& output);
    bool IsBlockingEnabled();
    bool ReportStats();
    void LoadConfig(const std::string& previousToken, const std::string& currentToken);
    void LoadConfigFromEnvironment();
    void LoadConfigWithTokenFromPHPSetToken(const std::string& tokenFromMiddleware);
    void RequestShutdown();


    ~RequestProcessorInstance();
};

extern RequestProcessor requestProcessor;
