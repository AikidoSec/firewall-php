#pragma once

#include <stack>

class RequestCache {
   public:
    std::string userId;
    std::string userName;
    std::string rateLimitGroup;
    std::string outgoingRequestUrl;
    std::string outgoingRequestRedirectUrl;

    RequestCache() = default;

/*
    Reset helper:

    This function re-initialize the cache structs to their default state instead
    of reallocating them. The PHP extension code runs inside long-lived PHP/Apache/FPM
    processes that handle many HTTP requests. Because these cache objects live for
    the lifetime of the process, we must explicitly reset them so that no state
    from one request or event can leak into the next.
*/
    void Reset();
};

class EventCache {
   public:
    std::string functionName;
    std::string moduleName;

    std::string filename;
    std::string filename2;

    std::string cmd;

    std::string outgoingRequestUrl;
    std::string outgoingRequestEffectiveUrl;
    std::string outgoingRequestPort;
    std::string outgoingRequestResolvedIp;
    std::string outgoingRequestEffectiveUrlPort;

    std::string sqlQuery;
    std::string sqlDialect;

    std::string paramMatcherParam;
    std::string paramMatcherRegex;

    EventCache() = default;
    void Reset();
};

class EventCacheStack {
   private:
    std::stack<EventCache> contexts;
   public:
    void Push();
    void Pop();
    EventCache& Top();
    bool Empty();
};

// RAII wrapper for automatic push/pop of event context
class ScopedEventContext {
   public:
    ScopedEventContext();
    ~ScopedEventContext();
};

