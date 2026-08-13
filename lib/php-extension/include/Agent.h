#pragma once

typedef GoUint8 (*AgentInitFn)(GoString initJson);
typedef void (*AgentUninitFn)();

class Agent {
   private:
    bool Start(std::string aikidoAgentPath);

   public:
    Agent() = default;
    ~Agent() = default;

    bool Init();
    void Uninit();
};
