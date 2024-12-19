#pragma once

typedef GoUint8 (*RequestProcessorInitFn)(GoString initJson);
typedef GoUint8 (*RequestProcessorContextInitFn)(ContextCallback);
typedef GoUint8 (*RequestProcessorConfigUpdateFn)(GoString initJson);
typedef char* (*RequestProcessorOnEventFn)(GoInt eventId);
typedef int (*RequestProcessorGetBlockingModeFn)();
typedef void (*RequestProcessorReportStats)(GoString, GoInt32, GoInt32, GoInt32, GoInt32, GoInt32, GoSlice);
typedef void (*RequestProcessorUninitFn)();

class RequestProcessor {
   private:
    bool initFailed = false;
    bool requestInitialized = false;
    bool configReloaded = false;
    void* libHandle = nullptr;
    uint64_t numberOfRequests = 0;
    RequestProcessorContextInitFn requestProcessorContextInitFn = nullptr;
    RequestProcessorConfigUpdateFn requestProcessorConfigUpdateFn = nullptr;
    RequestProcessorOnEventFn requestProcessorOnEventFn = nullptr;
    RequestProcessorGetBlockingModeFn requestProcessorGetBlockingModeFn = nullptr;
    RequestProcessorReportStats requestProcessorReportStatsFn = nullptr;
    RequestProcessorUninitFn requestProcessorUninitFn = nullptr;

   private:
    std::string GetInitData();
    bool ContextInit();
    void SendPreRequestEvent();
    void SendPostRequestEvent();

   public:
    RequestProcessor() = default;

    bool Init();
    bool RequestInit();
    bool SendEvent(EVENT_ID eventId, std::string& output);
    bool IsBlockingEnabled();
    bool ReportStats();
    void LoadConfigOnce();
    void RequestShutdown();
    void Uninit();

    ~RequestProcessor();
};

extern RequestProcessor requestProcessor;
