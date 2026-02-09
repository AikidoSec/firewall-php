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

