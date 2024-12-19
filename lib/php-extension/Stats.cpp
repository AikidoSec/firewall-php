#include "Includes.h"

std::unordered_map<std::string, SinkStats> stats;

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
    SinkStats& sinkStats = stats[this->key];
    sinkStats.timings.push_back(this->duration);
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