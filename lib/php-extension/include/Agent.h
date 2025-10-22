#pragma once

typedef GoUint8 (*AgentInitFn)(GoString initJson);
typedef void (*AgentUninitFn)();

class Agent {
   private:
    std::string socketPath;

    bool SocketFileExists();
    pid_t GetPID(const std::string& aikidoAgentPath);
    bool RemoveSocketFiles();

    bool Start(std::string aikidoAgentPath, std::string token);
    bool SpawnDetached(std::string aikidoAgentPath, std::string token);

   public:
    Agent() = default;
    ~Agent() = default;

    bool Init();
    void Uninit();
};
