#pragma once

#include <cstdint>
#include <string>
#include <unordered_map>

struct CurlTransfer {
    std::string body;
    bool discarded = false;
    const zend_function* callback = nullptr;
};

struct CurlCapture {
    void Reset();

    std::unordered_map<uint32_t, CurlTransfer> transfers;
    std::unordered_map<uint32_t, const zend_function*> callbacks;
};
