#pragma once

enum AIKIDO_LOG_LEVEL {
    AIKIDO_LOG_LEVEL_DEBUG,
    AIKIDO_LOG_LEVEL_INFO,
    AIKIDO_LOG_LEVEL_WARN,
    AIKIDO_LOG_LEVEL_ERROR
};

#if defined(ZEND_DEBUG)
#define AIKIDO_LOG_DEBUG(format, ...) printf(format, ##__VA_ARGS__)
#else
/* Disable debugging logs for production builds */
#define AIKIDO_LOG_DEBUG(format, ...) printf(format, ##__VA_ARGS__)
#endif

#define AIKIDO_LOG_INFO(format, ...) printf(format, ##__VA_ARGS__)

#define AIKIDO_LOG_WARN(format, ...) printf(format, ##__VA_ARGS__)
#define AIKIDO_LOG_ERROR(format, ...) printf(format, ##__VA_ARGS__)

class Log {
   private:
    std::string logFilePath;
    FILE* logFile = nullptr;

   public:
    Log() = default;
    ~Log() = default;

    void Init();
    void Uninit();

    void Write(AIKIDO_LOG_LEVEL level, const char* format, ...);

    static std::string ToString(AIKIDO_LOG_LEVEL level);

    static AIKIDO_LOG_LEVEL ToLevel(std::string level);
};

class LogScopedUninit {
   public:
    LogScopedUninit() = default;
    ~LogScopedUninit();
};