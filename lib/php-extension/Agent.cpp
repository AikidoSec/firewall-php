#include "Includes.h"

/*
 * Agent startup must be safe in two execution contexts:
 *
 * - Concurrent initialization across processes: overlapping PHP-FPM masters,
 *   Apache generations, `php -S` processes, or FrankenPHP instances may call
 *   Agent::Init() at the same time. Each caller starts its own launcher. On the
 *   Go side, the versioned runtime-directory lock admits one worker, while every
 *   launcher waits for the shared Unix socket to become ready.
 * - Initialization inside a multithreaded host: FrankenPHP may call Agent::Init()
 *   while other host threads exist. posix_spawn() starts the launcher as a fresh
 *   executable without running extension code in a manually forked child, which
 *   could inherit locks or runtime state from threads absent in that child.
 *
 * PHP waits for and reaps only the short-lived launcher. The long-lived worker
 * is shared and independent of any individual PHP process. PHP CLI intentionally
 * skips Agent startup.
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

    // PHP must reap the child it starts, but it cannot wait for the long-lived
    // Agent worker. It therefore waits only for the launcher, which exits after
    // it knows that the shared worker is ready or that startup failed.
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
