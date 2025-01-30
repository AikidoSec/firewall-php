#include "Includes.h"

std::unordered_map<std::string, SinkStats> stats;

std::chrono::high_resolution_clock::time_point currentRequestStart = std::chrono::high_resolution_clock::time_point{};

uint64_t totalOverheadForCurrentRequest = 0;

inline void AddToStats(const std::string& key, uint64_t duration) {
    SinkStats& sinkStats = stats[key];
    sinkStats.timings.push_back(duration);
}

inline void AddRequestTotalToStats() {
    if (currentRequestStart == std::chrono::high_resolution_clock::time_point{}) {
        return;
    }
    uint64_t totalOverhead = std::chrono::duration_cast<std::chrono::nanoseconds>(std::chrono::high_resolution_clock::now() - currentRequestStart).count();
    AddToStats("request_total", totalOverhead);
    currentRequestStart = std::chrono::high_resolution_clock::time_point{};
}

inline void AddRequestTotalOverheadToStats() {
    AddToStats("request_total_overhead", totalOverheadForCurrentRequest);
    totalOverheadForCurrentRequest = 0;
}

ScopedTimer::ScopedTimer() {
    this->Start();
}

ScopedTimer::ScopedTimer(std::string key) : key(key) {
    this->Start();
}

void ScopedTimer::SetSink(std::string key) {
    this->key = key;
}

void ScopedTimer::Start() {
    this->start = std::chrono::high_resolution_clock::now();
    if (this->key == "request_init") {
        currentRequestStart = this->start;
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
    if (this->key.empty()) {
        return;
    }
    this->Stop();
    totalOverheadForCurrentRequest += this->duration;
    if (key == "request_shutdown") {
        AddRequestTotalOverheadToStats();
        AddRequestTotalToStats();
    }
    AddToStats(this->key, this->duration);
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
