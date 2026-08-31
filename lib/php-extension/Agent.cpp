#include "Includes.h"

/*
 * Start a fresh executable because this may run in a multithreaded host, where
 * work after fork could deadlock on inherited runtime locks. PHP waits for only
 * the short-lived launcher; the lock-selected worker remains shared across PHP
 * processes after the launcher exits.
 */
bool Agent::Init() {
    std::string aikidoAgentPath = "/opt/aikido-" + std::string(PHP_AIKIDO_VERSION) + "/aikido-agent";

    AIKIDO_LOG_INFO("Starting Aikido Agent...\n");

    char* argv[] = {
        const_cast<char*>(aikidoAgentPath.c_str()),
        nullptr
    };

    pid_t agentPid;
    int status = posix_spawn(&agentPid, aikidoAgentPath.c_str(), nullptr, nullptr, argv, nullptr);
    if (status != 0) {
        AIKIDO_LOG_ERROR("Failed to start Aikido Agent process: %s\n", strerror(status));
        return false;
    }

    int waitStatus;
    while (waitpid(agentPid, &waitStatus, 0) == -1) {
        if (errno != EINTR) {
            AIKIDO_LOG_ERROR("Failed to wait for Aikido Agent launcher: %s\n", strerror(errno));
            return false;
        }
    }

    return WIFEXITED(waitStatus) && WEXITSTATUS(waitStatus) == 0;
}

void Agent::Uninit() {
    // The Agent worker is shared by all PHP processes, not owned by this one.
}
