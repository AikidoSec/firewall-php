#include "Includes.h"

ACTION aikido_execute_output_throw(json event) {
    std::string message = event["message"].get<std::string>();
    int code = event["code"].get<int>();
    zend_throw_exception(zend_exception_get_default(), message.c_str(), code);
    return BLOCK;
}

ACTION aikido_execute_output_exit(json event) {
    int _response_code = event["response_code"].get<int>();
    std::string _message = event["message"].get<std::string>();

    aikido_call_user_function("ob_clean");
    aikido_call_user_function("header_remove");
    aikido_call_user_function("http_response_code", _response_code);
    aikido_call_user_function_one_param("header", "Content-Type: text/plain");
    aikido_call_user_function_one_param("echo", _message);

#if PHP_VERSION_ID >= 80000
    aikido_call_user_function("exit");
#endif
    
    return EXIT;
}

ACTION aikido_execute_output(json event) {
    if (event["action"] == "throw") return aikido_execute_output_throw(event);
	if (event["action"] == "exit")  return aikido_execute_output_exit(event);
	return CONTINUE;
}