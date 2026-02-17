#pragma once

#include "Includes.h"

enum ACTION_STATUS {
    CONTINUE,
    BLOCK,
    EXIT,
    WARNING_MESSAGE
};

class Action {
    private:
        bool block = false;
        std::string type;
        std::string trigger;
        std::string description;
        std::string ip;
        std::string userAgent;

    private:
        ACTION_STATUS executeThrow(json &event);

        ACTION_STATUS executeExit(json &event);

        ACTION_STATUS executeStore(json &event);

        ACTION_STATUS executeWarningMessage(json &event);

        ACTION_STATUS executeBypassIp(json &event);

    public:
        Action() = default;
        ~Action() = default;

        ACTION_STATUS Execute(std::string &event);
        bool IsDetection(std::string &event);
        bool IsIdorViolation(std::string &event);

        void Reset();

        bool Exit();
        bool Block();
        char* Type();
        char* Trigger();
        char* Description();
        char* Ip();
        char* UserAgent();
};

extern Action action;
