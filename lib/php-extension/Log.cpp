#include "Includes.h"

void Log::Init() {
    if (!AIKIDO_GLOBAL(disk_logs)) {
        return;
    }

    this->logFilePath = "/var/log/aikido-" + std::string(PHP_AIKIDO_VERSION) + "/aikido-extension-php-" + GetDateTime() + "-" + std::to_string(getpid()) + ".log";
    this->logFile = fopen(this->logFilePath.c_str(), "w");
    AIKIDO_LOG_INFO("Opened log file %s!\n", this->logFilePath.c_str());
}

void Log::Uninit() {
    if (!this->logFile) {
        return;
    }
    AIKIDO_LOG_INFO("Closed log file %s!\n", this->logFilePath.c_str());
    fclose(this->logFile);
    this->logFile = nullptr;
}

void Log::Write(AIKIDO_LOG_LEVEL level, const char* format, ...) {
    if (level < AIKIDO_GLOBAL(log_level)) {
        return;
    }

    FILE* logFile = stdout;
    if (AIKIDO_GLOBAL(disk_logs)) {
        logFile = this->logFile;
    }

    if (!this->logFile) {
        return;
    }

    fprintf(logFile, "[AIKIDO][%s][%d][%s] ", ToString(level).c_str(), getpid(), GetTime().c_str());

    va_list args;
    va_start(args, format);
    vfprintf(logFile, format, args);
    va_end(args);

    fflush(logFile);
}

std::string Log::ToString(AIKIDO_LOG_LEVEL level) {
    switch (level) {
        case AIKIDO_LOG_LEVEL_DEBUG:
            return "DEBUG";
        case AIKIDO_LOG_LEVEL_INFO:
            return "INFO";
        case AIKIDO_LOG_LEVEL_WARN:
            return "WARN";
        case AIKIDO_LOG_LEVEL_ERROR:
            return "ERROR";
    }
    return "UNKNOWN";
}

AIKIDO_LOG_LEVEL Log::ToLevel(std::string level) {
    if (level == "ERROR") {
        return AIKIDO_LOG_LEVEL_ERROR;
    }
    if (level == "WARN") {
        return AIKIDO_LOG_LEVEL_WARN;
    }
    if (level == "INFO") {
        return AIKIDO_LOG_LEVEL_INFO;
    }
    if (level == "DEBUG") {
        return AIKIDO_LOG_LEVEL_DEBUG;
    }
    return AIKIDO_LOG_LEVEL_ERROR;
}

LogScopedUninit::~LogScopedUninit() {
    AIKIDO_GLOBAL(logger).Uninit();
}
