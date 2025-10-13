#pragma once

typedef GoUint8 (*AgentInitFn)(GoString initJson);
typedef void (*AgentUninitFn)();

class Agent {
   private:
    std::string socketPath;

    std::string GetInitData();
    std::string GetSocketPath();

    bool Start(std::string aikidoAgentPath, std::string initData, std::string token);
    bool SpawnDetached(std::string aikidoAgentPath, std::string initData, std::string token);

   public:
    Agent() = default;
    ~Agent() = default;

    bool Init();
    void Uninit();
};
