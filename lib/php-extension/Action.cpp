#include "Includes.h"

#define CONTENT_TYPE_HEADER "Content-Type: text/plain"

ACTION_STATUS Action::executeThrow(json &event) {
    int _code = event["code"].get<int>();
    std::string _message = event["message"].get<std::string>();
    SG(sapi_headers).http_response_code = _code;
    zend_throw_exception(zend_exception_get_default(), _message.c_str(), _code);
    return BLOCK;
}

ACTION_STATUS Action::executeExit(json &event) {
    int _response_code = event["response_code"].get<int>();
    std::string _message = event["message"].get<std::string>();

    // CallPhpFunction("ob_clean");
    CallPhpFunction("header_remove");
    SG(sapi_headers).http_response_code = _response_code;
    
    sapi_header_line ctr = {0};
    ctr.line = CONTENT_TYPE_HEADER;
    ctr.line_len = sizeof(CONTENT_TYPE_HEADER) - 1;
    ctr.response_code = 0;
    sapi_header_op(SAPI_HEADER_REPLACE, &ctr);
    
    CallPhpEcho(_message);
    CallPhpExit();
    return EXIT;
}

ACTION_STATUS Action::executeStore(json &event) {
    block = true;
    type = event["type"];
    trigger = event["trigger"];
    description = event["description"];
    if (trigger == "ip") {
        ip = event["ip"];
    }
    if (trigger == "user-agent") {
        userAgent = event["user-agent"];
    }
    return CONTINUE;
}

ACTION_STATUS Action::Execute(std::string &event) {
    if (event.empty()) {
        return CONTINUE;
    }

    json eventJson = json::parse(event);
    if (eventJson.empty()) {
        return CONTINUE;
    }
    std::string actionType = eventJson["action"];

    if (actionType == "throw") {
        return executeThrow(eventJson);
    } else if (actionType == "exit") {
        return executeExit(eventJson);
    } else if (actionType == "store") {
        return executeStore(eventJson);
    }
    return CONTINUE;
}

bool Action::IsDetection(std::string &event) {
    return !event.empty();
}

void Action::Reset() {
    block = false;
    type = "";
    trigger = "";
    description = "";
    ip = "";
    userAgent = "";
}

bool Action::Exit() {
    return exit;
}

bool Action::Block() {
    return block;
}

char *Action::Type() {
    return (char *)type.c_str();
}

char *Action::Trigger() {
    return (char *)trigger.c_str();
}

char *Action::Description() {
    return (char *)description.c_str();
}

char *Action::Ip() {
    return (char *)ip.c_str();
}

char *Action::UserAgent() {
    return (char *)userAgent.c_str();
}
