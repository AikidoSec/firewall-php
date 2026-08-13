#include "Includes.h"

bool Agent::Start(std::string aikidoAgentPath) {
    // The safety comes from crossing an exec boundary before detaching, not
    // from the launcher being written in Go. This extension may run inside a
    // multithreaded host such as FrankenPHP. A forked child would retain locks
    // and runtime state from threads that no longer exist in that child, making
    // extension work between fork and exec unsafe. posix_spawn instead loads a
    // clean launcher process without running our extension code in such a child.
    posix_spawnattr_t attr;
    posix_spawnattr_init(&attr);

    char* argv[] = {
        const_cast<char*>(aikidoAgentPath.c_str()),
        nullptr
    };

    pid_t agentPid;
    int status = posix_spawn(&agentPid, aikidoAgentPath.c_str(), nullptr, &attr, argv, nullptr);
    posix_spawnattr_destroy(&attr);
    if (status != 0) {
        AIKIDO_LOG_ERROR("Failed to start Aikido Agent process: %s\n", strerror(status));
        return false;
    }

    // Waiting for the real Agent would block PHP startup, while not waiting
    // could leave its exit status unreaped for the lifetime of this PHP process.
    // PHP therefore owns and reaps only the short-lived launcher. The launcher
    // starts the detached singleton worker and exits, so the worker is not a
    // long-lived direct child that this extension must wait for.
    int waitStatus;
    while (waitpid(agentPid, &waitStatus, 0) == -1) {
        if (errno != EINTR) {
            AIKIDO_LOG_ERROR("Failed to wait for Aikido Agent launcher: %s\n", strerror(errno));
            return false;
        }
    }

    return WIFEXITED(waitStatus) && WEXITSTATUS(waitStatus) == 0;
}

bool Agent::Init() {
    std::string aikidoAgentPath = "/opt/aikido-" + std::string(PHP_AIKIDO_VERSION) + "/aikido-agent";

    AIKIDO_LOG_INFO("Starting Aikido Agent...\n");

    if (!this->Start(aikidoAgentPath)) {
        AIKIDO_LOG_ERROR("Failed to start Aikido Agent!\n");
        return false;
    }

    return true;
}

void Agent::Uninit() {
    // The Agent worker is shared by all PHP processes, not owned by this one.
}
