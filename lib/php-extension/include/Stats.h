#pragma once

class ScopedTimer {
   private:
    std::string key;
    std::string kind;
    std::chrono::high_resolution_clock::time_point start;
    uint64_t duration = 0;
    
   public:
    ScopedTimer();
    ScopedTimer(std::string key, std::string kind);
    void SetSink(std::string key, std::string kind);
    void Start();
    void Stop();
    ~ScopedTimer();
};

class SinkStats {
    public:
     std::string kind;
     uint64_t attacksDetected = 0;
     uint64_t attacksBlocked = 0;
     uint64_t interceptorThrewError = 0;
     uint64_t withoutContext = 0;
     std::vector<int64_t> timings;

    void IncrementAttacksDetected();
    void IncrementAttacksBlocked();
    void IncrementInterceptorThrewError();
    void IncrementWithoutContext();
};

extern std::unordered_map<std::string, SinkStats> stats;
