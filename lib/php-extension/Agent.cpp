#include "Includes.h"

vector<pid_t> Agent::GetPIDsFromRunningProcesses(const std::string& aikidoAgentPath) {
    vector<pid_t> agentPIDs;

    DIR *dirTree = opendir("/proc");
    if (dirTree) {
        struct dirent *dirTreeEntry;
        while (dirTreeEntry = readdir(dirTree)) {
            int pid = atoi(dirTreeEntry->d_name);
            if (pid > 0) {
                string cmdPath = string("/proc/") + dirTreeEntry->d_name + "/cmdline";
                ifstream cmdFile(cmdPath.c_str());
                string cmdLine;
                getline(cmdFile, cmdLine);
                if (!cmdLine.empty() && cmdLine.find(aikidoAgentPath) != string::npos) {
                    agentPIDs.push_back(pid);
                }
            }
        }
        closedir(dirTree);
    }

    return agentPIDs;
}

pid_t Agent::GetPIDFromFile(const std::string& aikidoAgentPidPath) {
    std::ifstream pidFile(aikidoAgentPidPath);
    if (pidFile.is_open()) {
        int pid;
        pidFile >> pid;
        return pid;
    }
    return -1;
}

bool Agent::Start(std::string aikidoAgentPath) {
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

    AIKIDO_LOG_INFO("Aikido Agent started (pid: %d)!\n", agentPid);
    return true;
}

bool Agent::SpawnDetached(std::string aikidoAgentPath) {
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
        this->Start(aikidoAgentPath);
        _exit(0);
    }

    // Parent process
    int wstatus;
    waitpid(pid, &wstatus, 0);
    return WIFEXITED(wstatus) && WEXITSTATUS(wstatus) == 0;
}

bool Agent::RemoveSocketFile(const std::string& aikidoAgentSocketPath) {
    if (!RemoveFile(aikidoAgentSocketPath)) {
        AIKIDO_LOG_WARN("Failed to remove socket file \"%s\"!\n", aikidoAgentSocketPath.c_str());
        return false;
    }
    AIKIDO_LOG_INFO("Successfully removed socket file \"%s\"!\n", aikidoAgentSocketPath.c_str());
    return true;
}

void Agent::KillProcesses(std::vector<pid_t>& pids) {
    for (pid_t pid : pids) {
        if (kill(pid, SIGTERM) != 0) {
            AIKIDO_LOG_WARN("Failed to terminate Aikido Agent process %d!\n", pid);
        } else {
            AIKIDO_LOG_INFO("Successfully terminated Aikido Agent process %d!\n", pid);
        }
    }
}

bool Agent::IsRunning(const std::string& aikidoAgentPath, const std::string& aikidoAgentSocketPath) {
    if (!FileExists(aikidoAgentSocketPath)) {
        AIKIDO_LOG_INFO("No socket file found!\n");
        return false;
    } 
    
    AIKIDO_LOG_INFO("Found socket file \"%s\" on disk! Checking if Aikido Agent process is running...\n", aikidoAgentSocketPath.c_str());

    std::string aikidoAgentPidPath = "/run/aikido-" + std::string(PHP_AIKIDO_VERSION) + "/aikido-agent.pid";
    pid_t agentPIDFromFile = this->GetPIDFromFile(aikidoAgentPidPath);
    vector<pid_t> agentPIDsFromRunningProcesses = this->GetPIDsFromRunningProcesses(aikidoAgentPath);
    if (agentPIDFromFile == -1 || 
        agentPIDsFromRunningProcesses.size() != 1 || 
        agentPIDFromFile != agentPIDsFromRunningProcesses[0]) {
        AIKIDO_LOG_INFO("Aikido Agent not running: PID file %d, running process PIDs %s!\n", agentPIDFromFile, agentPIDsFromRunningProcesses.size() > 0 ? to_string(agentPIDsFromRunningProcesses[0]).c_str() : "-1");
        this->KillProcesses(agentPIDsFromRunningProcesses);
        this->RemoveSocketFile(aikidoAgentSocketPath);
        return false;
    }

    return true;
}

bool Agent::Init() {
    std::string aikidoAgentPath = "/opt/aikido-" + std::string(PHP_AIKIDO_VERSION) + "/aikido-agent";
    std::string aikidoAgentSocketPath = "/run/aikido-" + std::string(PHP_AIKIDO_VERSION) + "/aikido-agent.sock";

    if (this->IsRunning(aikidoAgentPath, aikidoAgentSocketPath)) {
        AIKIDO_LOG_INFO("Aikido Agent is already running! Skipping init...\n");
        return true;
    }

    AIKIDO_LOG_INFO("Starting Aikido Agent...\n");

    if (!this->SpawnDetached(aikidoAgentPath)) {
        AIKIDO_LOG_ERROR("Failed to spawn Aikido Agent in detached mode!\n");
        return false;
    }

    return true;
}

void Agent::Uninit() {
    // Nothing to do, Aikido Agent will terminate by itself
}
