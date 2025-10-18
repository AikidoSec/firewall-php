#include "Includes.h"

std::string Agent::GetInitData() {
    json initData = {{"socket_path", AIKIDO_GLOBAL(socket_path)},
                     {"log_level", AIKIDO_GLOBAL(log_level_str)},
                     {"disk_logs", AIKIDO_GLOBAL(disk_logs)}};

    // Remove invalid UTF8 characters (normalize)
    // https://json.nlohmann.me/api/basic_json/dump/
    return NormalizeAndDumpJson(initData);
}

std::string Agent::GetSocketPath() {
    std::string aikidoRunFolder = "/var/run/aikido-" + std::string(PHP_AIKIDO_VERSION);
    std::string socketPath = "";
    
    DIR* dir;
    struct dirent* ent;
    if ((dir = opendir(aikidoRunFolder.c_str())) != NULL) {
        while ((ent = readdir(dir)) != NULL) {
            std::string filename(ent->d_name);
            if (filename.find(".sock") != std::string::npos) {
                socketPath = aikidoRunFolder + "/" + filename;
                break;
            }
        }
        closedir(dir);
    } else {
        AIKIDO_LOG_WARN("Failed to open directory %s!\n", aikidoRunFolder.c_str());
    }

    return socketPath;
}

bool Agent::Start(std::string aikidoAgentPath, std::string initData, std::string token) {
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

    pid_t agentPid;
    int status = posix_spawn(&agentPid, aikidoAgentPath.c_str(), nullptr, &attr, argv, envp);
    posix_spawnattr_destroy(&attr);
    if (status != 0) {
        AIKIDO_LOG_ERROR("Failed to start Aikido Agent process: %s\n", strerror(status));
        return false;
    }

    AIKIDO_LOG_INFO("Aikido Agent started (pid: %d)!\n", agentPid);
    return true;
}

/*
bool Agent::SpawnDetached(std::string aikidoAgentPath, std::string initData, std::string token) {
    pid_t pid = fork();
    if ( pid < 0 ) {
        AIKIDO_LOG_ERROR("Failed to fork first child: %s\n", strerror(errno));
        return false;
    }
    
    if (pid == 0) {
        // Child process
        pid_t pid2 = fork();
        if (pid2 < 0) {
            AIKIDO_LOG_ERROR("Failed to fork second child: %s\n", strerror(errno));
            return false;
        }

        if (pid2 > 0) {
            // First child exits here
            _exit(0);
        }

        // Grandchild process
        // Now re-parented to init/systemd after parent exits
        // Create new session (detach from controlling terminal)
        if (setsid() < 0) {
            AIKIDO_LOG_ERROR("Failed to setsid: %s\n", strerror(errno));
            return false;
        }

        this->Start(aikidoAgentPath, initData, token);

        // We can _exit(0) here, since the spawned process is independent
        _exit(0);
    }
    
    // Parent waits for the first child to exit
    int wstatus;
    waitpid(pid, &wstatus, 0);
    return true;
}
*/

bool Agent::SpawnDetached(std::string aikidoAgentPath, std::string initData, std::string token) {
    pid_t pid = fork();
    if (pid < 0) {
        AIKIDO_LOG_ERROR("Failed to fork: %s\n", strerror(errno));
        return false;
    }

    if (pid == 0) {
        // Child process
        if (daemon(0, 0) != 0) {
            AIKIDO_LOG_ERROR("Failed to daemonize: %s\n", strerror(errno));
            _exit(1);
        }
        this->Start(aikidoAgentPath, initData, token);
        _exit(0);
    }

    // Parent process
    int wstatus;
    waitpid(pid, &wstatus, 0);
    return WIFEXITED(wstatus) && WEXITSTATUS(wstatus) == 0;
    //return true;
}

bool Agent::Init() {
    AIKIDO_GLOBAL(socket_path) = this->GetSocketPath();
    if (!AIKIDO_GLOBAL(socket_path).empty()) {
        AIKIDO_LOG_WARN("Aikido Agent already running on socket %s!\n", AIKIDO_GLOBAL(socket_path).c_str());
        return true;
    }
    
    AIKIDO_GLOBAL(socket_path) = GenerateSocketPath();
    
    std::string aikidoAgentPath = "/opt/aikido-" + std::string(PHP_AIKIDO_VERSION) + "/aikido-agent";
    std::string initData = this->GetInitData();
    std::string token = std::string("AIKIDO_TOKEN=") + AIKIDO_GLOBAL(token);

    AIKIDO_LOG_INFO("Starting Aikido Agent (%s) with init data: %s\n", aikidoAgentPath.c_str(), initData.c_str());

    if (!this->SpawnDetached(aikidoAgentPath, initData, token)) {
        AIKIDO_LOG_ERROR("Failed to spawn Aikido Agent in detached mode!\n");
        return false;
    }

    return true;
}

void Agent::Uninit() {
    // Nothing to do, Aikido Agent will terminate by itself
}
