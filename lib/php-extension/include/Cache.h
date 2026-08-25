#pragma once

#include <stack>
#include <string>
#include <unordered_map>
#include <vector>

class RequestCache {
   public:
    std::string userId;
    std::string userName;
    std::string rateLimitGroup;
    std::string outgoingRequestUrl;
    std::string outgoingRequestRedirectUrl;

    bool idorProtectionEnabled = false;
    std::string idorTenantColumnName;
    std::vector<std::string> idorExcludedTables;
    bool tenantIdSet = false;
    std::string tenantId;
    // Depth, not bool: allows nested without_idor_protection() calls.
    int idorIgnoredDepth = 0;
    std::unordered_map<void*, int> idorIgnoredDepthByFiber;

    RequestCache() = default;

    void EnterIdorIgnoredScope();
    void LeaveIdorIgnoredScope();
    bool IsIdorIgnored() const;

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
    std::string sqlParams;

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
