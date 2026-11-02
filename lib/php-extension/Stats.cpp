#include "Includes.h"

inline void AddToStats(const std::string& key, const std::string& kind, uint64_t duration) {
    SinkStats& sinkStats = AIKIDO_GLOBAL(stats)[key];
    sinkStats.kind = kind;
    sinkStats.timings.push_back(duration);
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
