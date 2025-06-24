#pragma once

class Server {
   private:
    zval* GetServerVar();

   public:
    Server() = default;

    std::string GetVar(const char* var);

    std::string GetRoute();

    std::string GetStatusCode();

    std::string GetUrl();

    std::string GetBody();

    std::string GetQuery();

    std::string GetHeaders();

    bool IsHttps();

    ~Server() = default;
};

extern Server server;
