#pragma once

class Server {
   private:
    // Stores the constant string "_SERVER", representing the PHP global variable 
    // that is used to access data about the current request.
    zend_string* serverString = nullptr;

    zval* GetServerVar();

   public:
    Server() = default;

    void Init();

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
