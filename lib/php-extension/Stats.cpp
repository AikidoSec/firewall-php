#include "Includes.h"

inline void AddToStats(const std::string& key, const std::string& kind, uint64_t duration) {
    SinkStats& sinkStats = AIKIDO_GLOBAL(stats)[key];
    sinkStats.kind = kind;
    sinkStats.timings.push_back(duration);
}

inline void AddRequestTotalToStats() {
    if (AIKIDO_GLOBAL(currentRequestStart) == std::chrono::high_resolution_clock::time_point{}) {
        return;
    }
    uint64_t totalOverhead = std::chrono::duration_cast<std::chrono::nanoseconds>(std::chrono::high_resolution_clock::now() - AIKIDO_GLOBAL(currentRequestStart)).count();
    AddToStats("request_total", "request_op", totalOverhead);
    AIKIDO_GLOBAL(currentRequestStart) = std::chrono::high_resolution_clock::time_point{};
}

inline void AddRequestTotalOverheadToStats() {
    AddToStats("request_total_overhead", "request_op", AIKIDO_GLOBAL(totalOverheadForCurrentRequest));
    AIKIDO_GLOBAL(totalOverheadForCurrentRequest) = 0;
}

ScopedTimer::ScopedTimer() {
    this->Start();
}

ScopedTimer::ScopedTimer(std::string key, std::string kind) : key(key), kind(kind) {
    this->Start();
}

void ScopedTimer::SetSink(std::string key, std::string kind) {
    this->key = key;
    this->kind = kind;
}

void ScopedTimer::Start() {
    this->start = std::chrono::high_resolution_clock::now();
    if (this->key == "request_init") {
        AIKIDO_GLOBAL(currentRequestStart) = this->start;
    }
}

void ScopedTimer::Stop() {
    if (this->start == std::chrono::high_resolution_clock::time_point{}) {
        return;
    }
    this->duration += std::chrono::duration_cast<std::chrono::nanoseconds>(std::chrono::high_resolution_clock::now() - this->start).count();
    this->start = std::chrono::high_resolution_clock::time_point{};
}

ScopedTimer::~ScopedTimer() {
    if (this->key.empty() || this->kind.empty()) {
        return;
    }
    this->Stop();
    AIKIDO_GLOBAL(totalOverheadForCurrentRequest) += this->duration;
    if (key == "request_shutdown") {
        AddRequestTotalOverheadToStats();
        AddRequestTotalToStats();
    }
    AddToStats(this->key, this->kind, this->duration);
}

void SinkStats::IncrementAttacksDetected() {
    attacksDetected++;
}

void SinkStats::IncrementAttacksBlocked() {
    attacksBlocked++;
}

void SinkStats::IncrementInterceptorThrewError() {
    interceptorThrewError += 1;
}

void SinkStats::IncrementWithoutContext() {
    withoutContext += 1;
}
