#include "Includes.h"

/*
 * Agent startup must be safe in two execution contexts:
 *
 * - Concurrent initialization across processes: overlapping PHP-FPM masters,
 *   Apache generations, `php -S` processes, or FrankenPHP instances may call
 *   Agent::Init() at the same time. The Go-side versioned runtime-directory lock
 *   ensures that their launchers converge on one shared worker.
 * - Initialization inside a multithreaded host: FrankenPHP may call Agent::Init()
 *   while other host threads exist. posix_spawn() starts the launcher as a fresh
 *   executable without running extension code in a manually forked child, which
 *   could inherit locks or runtime state from threads absent in that child.
 *
 * Startup follows this process sequence:
 *
 * PHP -> Agent in launcher mode (no arguments) -> Agent in worker mode (--agent-worker)
 *
 * The launcher exits after the shared Unix socket is ready or startup has failed.
 * PHP waits only for its launcher to exit before Agent::Init() returns. The worker
 * retains the runtime-directory lock and continues independently of the PHP
 * process that initiated startup. The launcher and worker modes share one
 * executable to avoid packaging a separate launcher.
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
