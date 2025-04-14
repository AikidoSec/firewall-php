#include "Includes.h"

std::string Agent::GetInitData() {
    json initData = {{"socket_path", AIKIDO_GLOBAL(socket_path)},
                     {"platform_name", AIKIDO_GLOBAL(sapi_name)},
                     {"platform_version", PHP_VERSION},
                     {"endpoint", AIKIDO_GLOBAL(endpoint)},
                     {"config_endpoint", AIKIDO_GLOBAL(config_endpoint)},
                     {"log_level", AIKIDO_GLOBAL(log_level_str)},
                     {"disk_logs", AIKIDO_GLOBAL(disk_logs)},
                     {"blocking", AIKIDO_GLOBAL(blocking)},
                     {"localhost_allowed_by_default", AIKIDO_GLOBAL(localhost_allowed_by_default)},
                     {"collect_api_schema", AIKIDO_GLOBAL(collect_api_schema)}};
    // Remove invalid UTF8 characters (normalize)
    // https://json.nlohmann.me/api/basic_json/dump/
    return NormalizeAndDumpJson(initData);
}

bool Agent::Init() {
    if (this->agentPid != 0) {
        AIKIDO_LOG_WARN("Aikido Agent already running!\n");
        return true;
    }

    std::string aikidoAgentPath = "/opt/aikido-" + std::string(PHP_AIKIDO_VERSION) + "/aikido-agent";
    std::string initData = this->GetInitData();
    std::string token = std::string("AIKIDO_TOKEN=") + AIKIDO_GLOBAL(token);

    AIKIDO_LOG_INFO("Starting Aikido Agent with init data: %s\n", initData.c_str());

    posix_spawnattr_t attr;
    posix_spawnattr_init(&attr);

    char* argv[] = {
        const_cast<char*>(aikidoAgentPath.c_str()),
        const_cast<char*>(initData.c_str()),
        nullptr
    };

    char* envp[] = {
        const_cast<char*>(token.c_str()),
        nullptr
    };

    int status = posix_spawn(&this->agentPid, aikidoAgentPath.c_str(), nullptr, &attr, argv, envp);
    if (status != 0) {
        AIKIDO_LOG_ERROR("Failed to spawn Aikido Agent process: %s\n", strerror(status));
        posix_spawnattr_destroy(&attr);
        return false;
    }

    posix_spawnattr_destroy(&attr);

    AIKIDO_LOG_INFO("Aikido Agent started (pid: %d)!\n", this->agentPid);
    return true;
}

void Agent::Uninit() {
    if (this->agentPid == 0) {
        AIKIDO_LOG_WARN("Aikido Agent not running!\n");
        return;
    }

    AIKIDO_LOG_INFO("Stopping Aikido Agent...\n");
    kill(this->agentPid, SIGTERM);
    this->agentPid = 0;
    AIKIDO_LOG_INFO("Aikido Agent stopped!\n");
}
