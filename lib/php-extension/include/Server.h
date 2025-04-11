#pragma once

class Server {
   private:
    zend_string* serverString = nullptr;

    zval* GetServerVar();

   public:
    Server();

    std::string GetVar(const char* var);

    std::string GetRoute();

    std::string GetStatusCode();

    std::string GetUrl();

    std::string GetBody();

    std::string GetQuery();

    std::string GetHeaders();

    bool IsHttps();

    ~Server();
};

extern Server server;
