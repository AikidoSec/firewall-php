#pragma once

typedef void (*aikido_handler)(INTERNAL_FUNCTION_PARAMETERS, EVENT_ID &eventId);
typedef zend_op_array* (*aikido_file_compilation_handler)(zend_file_handle *file_handle, int type);

struct PHP_HANDLERS {
    aikido_handler handler;
    aikido_handler post_handler;
    zif_handler original_handler;
};

extern unordered_map<std::string, PHP_HANDLERS> HOOKED_FUNCTIONS;

class AIKIDO_METHOD_KEY {
   public:
    std::string class_name;
    std::string method_name;

    AIKIDO_METHOD_KEY(const std::string &class_name, const std::string &method_name) : class_name(class_name), method_name(method_name) {
    }

    bool operator==(const AIKIDO_METHOD_KEY &other) const {
        return class_name == other.class_name && method_name == other.method_name;
    }
};

class AIKIDO_METHOD_KEY_HASH {
   public:
    std::size_t operator()(const AIKIDO_METHOD_KEY &k) const {
        return std::hash<std::string>()(k.class_name) ^ (std::hash<std::string>()(k.method_name) << 1);
    }
};

extern unordered_map<AIKIDO_METHOD_KEY, PHP_HANDLERS, AIKIDO_METHOD_KEY_HASH> HOOKED_METHODS;

extern aikido_file_compilation_handler original_file_compilation_handler;

void HookFunctions();

void UnhookFunctions();

void HookMethods();

void UnhookMethods();

void HookFileCompilation();

void UnhookFileCompilation();
