#pragma once

typedef GoUint8 (*AgentInitFn)(GoString initJson);
typedef void (*AgentUninitFn)();

class Agent {
   private:
    pid_t GetPIDFromFile(const std::string& aikidoAgentPidPath);
    vector<pid_t> GetPIDsFromRunningProcesses(const std::string& aikidoAgentPath);
    
    bool RemoveSocketFile(const std::string& aikidoAgentSocketPath);
    void KillProcesses(std::vector<pid_t>& pids);
    
    bool IsRunning(const std::string& aikidoAgentPath, const std::string& aikidoAgentSocketPath);

    bool Start(std::string aikidoAgentPath, std::string token);
    bool SpawnDetached(std::string aikidoAgentPath, std::string token);

   public:
    Agent() = default;
    ~Agent() = default;

    bool Init();
    void Uninit();
};
