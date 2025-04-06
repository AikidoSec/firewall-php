#pragma once

class Request {
   private:
    zval* server = nullptr;

   public:
    Request() = default;

    bool IsServerVarLoaded();

    bool LoadServerVar();

    void UnloadServerVar();

    std::string GetVar(const char* var);

    std::string GetRoute();

    std::string GetStatusCode();

    std::string GetUrl();

    std::string GetBody();

    std::string GetQuery();

    std::string GetHeaders();

    bool IsHttps();

    ~Request() = default;
};

extern Request request;
