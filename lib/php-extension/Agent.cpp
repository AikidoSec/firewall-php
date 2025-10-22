#include "Includes.h"

std::string Agent::GetInitData() {
    json initData = {{"log_level", AIKIDO_GLOBAL(log_level_str)},
                     {"disk_logs", AIKIDO_GLOBAL(disk_logs)}};

    // Remove invalid UTF8 characters (normalize)
    // https://json.nlohmann.me/api/basic_json/dump/
    return NormalizeAndDumpJson(initData);
}

pid_t Agent::GetPID(const std::string& aikidoAgentPath) {
    int agentPID = -1;

    DIR *dirTree = opendir("/proc");
    if (dirTree) {
        struct dirent *dirTreeEntry;
        while (agentPID < 0 && (dirTreeEntry = readdir(dirTree))) {
            int currentPid = atoi(dirTreeEntry->d_name);
            if (currentPid > 0) {
                string cmdPath = string("/proc/") + dirTreeEntry->d_name + "/cmdline";
                ifstream cmdFile(cmdPath.c_str());
                string cmdLine;
                getline(cmdFile, cmdLine);
                if (!cmdLine.empty() && cmdLine.find(aikidoAgentPath) != string::npos) {
                    agentPID = currentPid;
                }
            }
        }
        closedir(dirTree);
    }

    return agentPID;
}

bool Agent::Start(std::string aikidoAgentPath, std::string token) {
    posix_spawnattr_t attr;
    posix_spawnattr_init(&attr);

    char* argv[] = {
        const_cast<char*>(aikidoAgentPath.c_str()),
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

bool Agent::SpawnDetached(std::string aikidoAgentPath, std::string token) {
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
        this->Start(aikidoAgentPath, token);
        _exit(0);
    }

    // Parent process
    int wstatus;
    waitpid(pid, &wstatus, 0);
    return WIFEXITED(wstatus) && WEXITSTATUS(wstatus) == 0;
}

bool Agent::Init() {
    std::string aikidoAgentPath = "/opt/aikido-" + std::string(PHP_AIKIDO_VERSION) + "/aikido-agent";
    std::string aikidoAgentSocketPath = "/run/aikido-" + std::string(PHP_AIKIDO_VERSION) + "/aikido-agent.sock";

    if (!std::filesystem::exists(aikidoAgentSocketPath)) {
        AIKIDO_LOG_INFO("Found socket file \"%s\" on disk! Checking if Aikido Agent is already running...\n", aikidoAgentSocketPath.c_str());
        pid_t agentPID = this->GetPID(aikidoAgentPath);
        if (agentPID != -1) {
            AIKIDO_LOG_INFO("Aikido Agent (PID: %d) already running on socket %s!\n", agentPID, aikidoAgentSocketPath.c_str());
            return true;    
        } else {
            AIKIDO_LOG_WARN("Aikido Agent is not running, but socket files exist! Recovering by removing old socket files...\n");
            
            if (!std::filesystem::remove(aikidoAgentSocketPath)) {
                AIKIDO_LOG_WARN("Failed to remove some socket files, will try to re-spawn Aikido Agent...\n");
            } else {
                AIKIDO_LOG_INFO("Successfully removed old socket files!\n");
            }
        }
    }
    
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
