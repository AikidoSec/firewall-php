#include "Includes.h"

Action action;

ACTION_STATUS Action::executeThrow(json &event) {
    int _code = event["code"].get<int>();
    std::string _message = event["message"].get<std::string>();
    zend_throw_exception(GetFirewallDefaultExceptionCe(), _message.c_str(), _code);
    CallPhpFunctionWithOneParam("http_response_code", _code);
    return BLOCK;
}

ACTION_STATUS Action::executeExit(json &event) {
    int _response_code = event["response_code"].get<int>();
    std::string _message = event["message"].get<std::string>();

    // CallPhpFunction("ob_clean");
    CallPhpFunction("header_remove");
    CallPhpFunctionWithOneParam("http_response_code", _response_code);
    CallPhpFunctionWithOneParam("header", "Content-Type: text/plain");
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

ACTION_STATUS Action::executeBypassIp(json &event) {
    SetIpBypassed();
    return CONTINUE;
}

ACTION_STATUS Action::executeWarningMessage(json &event) {
    std::string message = event["message"];
    php_printf("%s\n", message.c_str());
    return WARNING_MESSAGE;
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
    } else if (actionType == "warning_message") {
        return executeWarningMessage(eventJson);
    } else if (actionType == "bypassIp") {
        return executeBypassIp(eventJson);
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
