#pragma once

#include <string>

static const size_t MAX_AI_BODY_BYTES = 2 * 1024 * 1024;

struct AICapture {
    std::string stats;
};

bool AIPath(const std::string& url);
void CaptureAIResponse(const std::string& url, const std::string& body);
void ReportAICapture();
std::string GetAICaptureStats();
