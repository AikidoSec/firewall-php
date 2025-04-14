#pragma once

typedef GoUint8 (*AgentInitFn)(GoString initJson);
typedef void (*AgentUninitFn)();

class Agent {
   private:
    pid_t agentPid = 0;
    std::string GetInitData();

   public:
    Agent() = default;
    ~Agent() = default;

    bool Init();
    void Uninit();
};
