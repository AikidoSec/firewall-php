

class PhpLifecycle {
   private:
    pid_t mainPID = 0;

   public:
    PhpLifecycle() = default;
    ~PhpLifecycle() = default;

    void ModuleInit();

    void RequestInit();

    void RequestShutdown();

    void ModuleShutdown();

    void HookAll();

    void UnhookAll();
};
