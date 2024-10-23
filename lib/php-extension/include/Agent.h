#pragma once

typedef GoUint8 (*AgentInitFn)(GoString initJson);
typedef void (*AgentUninitFn)();

class Agent {
   private:
    void* libHandle = nullptr;

   private:
    std::string GetInitData();

   public:
    Agent() = default;

    bool Init();

    ~Agent();
};
