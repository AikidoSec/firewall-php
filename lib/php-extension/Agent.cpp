#include "Includes.h"

std::string Agent::GetInitData() {
    json initData = {{"token", AIKIDO_GLOBAL(token)},
                     {"socket_path", AIKIDO_GLOBAL(socket_path)},
                     {"platform_name", AIKIDO_GLOBAL(sapi_name)},
                     {"platform_version", PHP_VERSION},
                     {"endpoint", AIKIDO_GLOBAL(endpoint)},
                     {"config_endpoint", AIKIDO_GLOBAL(config_endpoint)},
                     {"log_level", AIKIDO_GLOBAL(log_level_str)},
                     {"blocking", AIKIDO_GLOBAL(blocking)},
                     {"localhost_allowed_by_default", AIKIDO_GLOBAL(localhost_allowed_by_default)},
                     {"collect_api_schema", AIKIDO_GLOBAL(collect_api_schema)}};
    // Remove invalid UTF8 characters (normalize)
    return initData.dump(-1, ' ', false, json::error_handler_t::ignore);
}

bool Agent::Init() {
    std::string aikido_agent_lib_handle_path =
        "/opt/aikido-" + std::string(PHP_AIKIDO_VERSION) + "/aikido-agent.so";
    this->libHandle = dlopen(aikido_agent_lib_handle_path.c_str(), RTLD_LAZY);
    if (!this->libHandle) {
        AIKIDO_LOG_ERROR("Error loading the Aikido Agent library from %s: %s!\n",
                         aikido_agent_lib_handle_path.c_str(), dlerror());
        return false;
    }

    AgentInitFn agentInitFn = (AgentInitFn)dlsym(this->libHandle, "AgentInit");
    if (!agentInitFn) {
        AIKIDO_LOG_ERROR(
            "Error loading symbol 'AgentInit' from the Aikido Agent library: %s!\n",
            dlerror());
        return false;
    }

    AIKIDO_LOG_INFO("Initializing Aikido Agent...\n");

    std::string initData = this->GetInitData();
    return agentInitFn(GoCreateString(initData));
}

void Agent::Uninit() {
    if (!this->libHandle) {
        return;
    }
    AgentUninitFn agentUninitFn =
        (AgentUninitFn)dlsym(this->libHandle, "AgentUninit");
    if (agentUninitFn) {
        AIKIDO_LOG_INFO("Uninitializing Aikido Agent library...\n");
        agentUninitFn();
        AIKIDO_LOG_INFO("Aikido Agent library uninitialized!\n");
    } else {
        AIKIDO_LOG_ERROR(
            "Error loading symbol 'AgentUninit' from Aikido Agent library: %s!\n",
            dlerror());
    }
    dlclose(this->libHandle);
    this->libHandle = nullptr;
}
