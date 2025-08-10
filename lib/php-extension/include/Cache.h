#pragma once

class RequestCache {
   public:
    std::string userId;
    std::string userName;

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

    std::string sqlQuery;
    std::string sqlDialect;

    unsigned int rateLimitGroup;

    EventCache() = default;
    void Reset();
};

extern RequestCache requestCache;
extern EventCache eventCache;
