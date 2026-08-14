#include "Includes.h"

/*
 * Agent startup overview
 *
 * PhpLifecycle::ModuleInit() calls Agent::Init() during PHP module startup for
 * every enabled server SAPI. The supported environments reach this code in
 * different ways:
 *
 * - PHP-FPM (whether behind Nginx, Apache, or Caddy): the FPM master normally
 *   initializes the extension before it forks workers. Separate or overlapping
 *   FPM masters sharing /run/aikido-<version> may start at the same time.
 * - Apache mod_php: PHP is initialized for an Apache server generation. An old
 *   and a new generation, or separate Apache instances, may overlap.
 * - PHP built-in server: each independently started `php -S` server initializes
 *   the extension. This is also the small process used by the startup stress test.
 * - FrankenPHP classic and worker modes: both use this module-startup path inside
 *   the same kind of multithreaded host. Their request lifecycles differ, but
 *   Agent startup does not.
 * - PHP CLI: PhpLifecycle::ModuleInit() intentionally skips Agent startup.
 *
 * C++ therefore does not inspect the PID file or socket and does not decide
 * whether an Agent already exists. Each server startup uses posix_spawn() to run
 * the short-lived Go launcher and waits only for that launcher. The launcher
 * selects one shared worker using the runtime-directory lock; simultaneous
 * launchers either start that worker or return success because another process
 * is already starting or running it. The worker owns the PID file, Unix socket,
 * and lock, and is not owned or stopped by any individual PHP process.
 */
bool Agent::Start(std::string aikidoAgentPath) {
    // posix_spawn() starts a new executable without asking this extension to run
    // code in a forked copy of the PHP host. That distinction is required for a
    // multithreaded host such as FrankenPHP.
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

    // PHP must reap the child it starts, but it cannot wait for the long-lived
    // Agent worker. It therefore waits only for the launcher, which starts the
    // worker and exits immediately.
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
