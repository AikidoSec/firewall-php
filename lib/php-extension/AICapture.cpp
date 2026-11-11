#include "Includes.h"
#include "AICapture.h"
#include <regex>

extern "C" {
#include "ext/standard/url.h"
}

static std::string Text(const json& object, const char* key) {
    auto it = object.find(key);
    return it != object.end() && it->is_string() ? it->get<std::string>() : "";
}

static int64_t Count(const json& object, const char* key) {
    auto it = object.find(key);
    if (it == object.end() || !it->is_number_integer()) {
        return 0;
    }
    if (it->is_number_unsigned() && it->get<uint64_t>() > 9007199254740991ULL) {
        return 0;
    }
    auto value = it->get<int64_t>();
    return value >= 0 && value <= 9007199254740991LL ? value : 0;
}

static json Object(const json& object, const char* key) {
    auto it = object.find(key);
    return it != object.end() && it->is_object() ? *it : json::object();
}

struct AIUsage {
    std::string model;
    int64_t input = 0;
    int64_t output = 0;
    int64_t cacheRead = 0;
    int64_t cacheWrite = 0;
};

static AIUsage ParseOpenAIChat(const std::smatch&, const json& body) {
    AIUsage call;
    call.model = Text(body, "model");
    auto usage = Object(body, "usage");
    call.input = Count(usage, "prompt_tokens");
    call.output = Count(usage, "completion_tokens");
    auto details = Object(usage, "prompt_tokens_details");
    call.cacheRead = Count(details, "cached_tokens");
    call.cacheWrite = Count(details, "cache_write_tokens");
    return call;
}

static AIUsage ParseOpenAIResponses(const std::smatch&, const json& body) {
    AIUsage call;
    call.model = Text(body, "model");
    auto usage = Object(body, "usage");
    call.input = Count(usage, "input_tokens");
    call.output = Count(usage, "output_tokens");
    auto details = Object(usage, "input_tokens_details");
    call.cacheRead = Count(details, "cached_tokens");
    call.cacheWrite = Count(details, "cache_write_tokens");
    return call;
}

static AIUsage ParseAnthropic(const std::smatch&, const json& body) {
    AIUsage call;
    call.model = Text(body, "model");
    auto usage = Object(body, "usage");
    call.input = Count(usage, "input_tokens");
    call.output = Count(usage, "output_tokens");
    call.cacheRead = Count(usage, "cache_read_input_tokens");
    call.cacheWrite = Count(usage, "cache_creation_input_tokens");
    return call;
}

static std::string BedrockModel(const std::smatch& match, const json& body) {
    std::string id = match[1].str();
    id.resize(php_raw_url_decode(&id[0], id.size()));
    return id.compare(0, 4, "arn:") == 0 ? Text(body, "model") : id;
}

static AIUsage ParseBedrockConverse(const std::smatch& match, const json& body) {
    AIUsage call;
    call.model = BedrockModel(match, body);
    auto usage = Object(body, "usage");
    call.input = Count(usage, "inputTokens");
    call.output = Count(usage, "outputTokens");
    call.cacheRead = Count(usage, "cacheReadInputTokens");
    call.cacheWrite = Count(usage, "cacheWriteInputTokens");
    return call;
}

static AIUsage ParseBedrockInvoke(const std::smatch& match, const json& body) {
    auto usage = Object(body, "usage");
    if (!usage.contains("input_tokens") && !usage.contains("output_tokens")) {
        return {};
    }
    auto call = ParseAnthropic(match, body);
    call.model = BedrockModel(match, body);
    return call;
}

static AIUsage ParseGemini(const std::smatch& match, const json& body) {
    AIUsage call;
    call.model = Text(body, "modelVersion");
    if (call.model.empty()) {
        call.model = match[1].str();
    }
    auto usage = Object(body, "usageMetadata");
    call.input = Count(usage, "promptTokenCount") + Count(usage, "toolUsePromptTokenCount");
    call.output = Count(usage, "candidatesTokenCount") + Count(usage, "thoughtsTokenCount");
    call.cacheRead = Count(usage, "cachedContentTokenCount");
    return call;
}

struct AIResponseParser {
    std::regex path;
    AIUsage (*parse)(const std::smatch&, const json&);
};

static const AIResponseParser AI_RESPONSE_PARSERS[] = {
    {std::regex(R"(/chat/completions$)"), ParseOpenAIChat},
    {std::regex(R"(/responses$)"), ParseOpenAIResponses},
    {std::regex(R"(/v1/messages$)"), ParseAnthropic},
    {std::regex(R"(/publishers/anthropic/models/[^/]+:rawPredict$)"), ParseAnthropic},
    {std::regex(R"(/models/([^/:]+):generateContent$)"), ParseGemini},
    {std::regex(R"(/model/([^/]+)/converse$)"), ParseBedrockConverse},
    {std::regex(R"(/model/([^/]+)/invoke$)"), ParseBedrockInvoke},
};

bool AIPath(const std::string& url) {
    std::string path = url.substr(0, url.find_first_of("?#"));
    for (const auto& parser : AI_RESPONSE_PARSERS) {
        if (std::regex_search(path, parser.path)) {
            return true;
        }
    }
    return false;
}

static json ParseCall(const std::string& url, const std::string& body) {
    auto limitDepth = [](int depth, json::parse_event_t, json&) {
        if (depth > 64) {
            throw std::runtime_error("AI response nesting limit");
        }
        return true;
    };
    json parsed = json::parse(body, limitDepth, false);
    if (!parsed.is_object()) {
        return nullptr;
    }
    php_url* uri = php_url_parse_ex(url.data(), url.size());
    if (!uri) {
        return nullptr;
    }
#if PHP_VERSION_ID >= 70300
    std::string scheme = uri->scheme ? ZSTR_VAL(uri->scheme) : "";
    std::string host = uri->host ? ZSTR_VAL(uri->host) : "";
    std::string path = uri->path ? ZSTR_VAL(uri->path) : "";
#else
    std::string scheme = uri->scheme ? uri->scheme : "";
    std::string host = uri->host ? uri->host : "";
    std::string path = uri->path ? uri->path : "";
#endif
    auto port = uri->port;
    php_url_free(uri);
    scheme = ToLowercase(scheme);
    host = ToLowercase(host);
    if ((scheme != "http" && scheme != "https") || host.empty()) {
        return nullptr;
    }
    std::string provider = scheme + "://" + host;
    if (port && !((scheme == "http" && port == 80) || (scheme == "https" && port == 443))) {
        provider += ":" + std::to_string(port);
    }
    AIUsage call;
    std::smatch match;
    for (const auto& parser : AI_RESPONSE_PARSERS) {
        if (std::regex_search(path, match, parser.path)) {
            call = parser.parse(match, parsed);
            break;
        }
    }
    if (call.model.empty() || call.model.size() > 256) {
        return nullptr;
    }
    return {{"provider", provider}, {"model", call.model}, {"input", call.input}, {"output", call.output}, {"cacheRead", call.cacheRead}, {"cacheWrite", call.cacheWrite}};
}

void CaptureAIResponse(const std::string& url, const std::string& body) {
    auto call = ParseCall(url, body);
    if (call.is_null()) {
        return;
    }
    auto& capture = AIKIDO_GLOBAL(aiCapture);
    capture.stats = call.dump();
    ReportAICapture();
    capture.stats.clear();
}

std::string GetAICaptureStats() {
    return AIKIDO_GLOBAL(aiCapture).stats;
}

void ReportAICapture() {
    std::string output;
    AIKIDO_GLOBAL(requestProcessorInstance).SendEvent(EVENT_AI_CALL, output);
}
