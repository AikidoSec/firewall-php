#pragma once

class Server {
   private:
    zval* GetServerVar();

   public:
    Server() = default;

    std::string GetVar(const char* var);

    std::string GetMethod();

    std::string getMethodFromQuery();

    std::string GetRoute();

    std::string GetStatusCode();

    std::string GetUrl();

    std::string GetBody();

    std::string GetQuery();

    std::string GetHeaders();

    std::string GetPost();

    bool IsHttps();

    ~Server() = default;
};

extern Server server;
