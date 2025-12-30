#pragma once

typedef void* (*CreateInstanceFn)(uint64_t threadId, bool isZTS);
typedef void (*DestroyInstanceFn)(uint64_t threadId);

// Updated typedefs with instance pointer as first parameter
typedef GoUint8 (*RequestProcessorInitFn)(void* instancePtr, GoString initJson);
typedef GoUint8 (*RequestProcessorContextInitFn)(void* instancePtr, ContextCallback);
typedef GoUint8 (*RequestProcessorConfigUpdateFn)(void* instancePtr, GoString initJson);
typedef char* (*RequestProcessorOnEventFn)(void* instancePtr, GoInt eventId);
typedef int (*RequestProcessorGetBlockingModeFn)(void* instancePtr);
typedef void (*RequestProcessorReportStats)(void* instancePtr, GoString, GoString, GoInt32, GoInt32, GoInt32, GoInt32, GoInt32, GoSlice);
typedef void (*RequestProcessorUninitFn)(void* instancePtr);

class RequestProcessor {
   private:
    bool initFailed = false;
    bool requestInitialized = false;
    void* libHandle = nullptr;
    void* requestProcessorInstance = nullptr; 
    uint64_t numberOfRequests = 0;
    
    // Function pointers to Go-exported functions
    CreateInstanceFn createInstanceFn = nullptr;
    DestroyInstanceFn destroyInstanceFn = nullptr;
    RequestProcessorInitFn requestProcessorInitFn = nullptr;
    RequestProcessorContextInitFn requestProcessorContextInitFn = nullptr;
    RequestProcessorConfigUpdateFn requestProcessorConfigUpdateFn = nullptr;
    RequestProcessorOnEventFn requestProcessorOnEventFn = nullptr;
    RequestProcessorGetBlockingModeFn requestProcessorGetBlockingModeFn = nullptr;
    RequestProcessorReportStats requestProcessorReportStatsFn = nullptr;
    RequestProcessorUninitFn requestProcessorUninitFn = nullptr;

   private:
    std::string GetInitData(const std::string& token = "");
    bool ContextInit();
    void SendPreRequestEvent();
    void SendPostRequestEvent();

   public:
    RequestProcessor() = default;

    bool Init();
    bool RequestInit();
    bool SendEvent(EVENT_ID eventId, std::string& output);
    bool IsBlockingEnabled();
    bool IsRequestInitialized() const { return requestInitialized; }
    bool ReportStats();
    void LoadConfig(const std::string& previousToken, const std::string& currentToken);
    void LoadConfigFromEnvironment();
    void LoadConfigWithTokenFromPHPSetToken(const std::string& tokenFromMiddleware);
    void RequestShutdown();
    void Uninit();

    ~RequestProcessor();
};
