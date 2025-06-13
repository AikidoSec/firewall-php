import requests
import time
import sys
from testlib import *

'''
1. Sets up the allowed IP address config for route '/test'. Checks that requests are blocked.
2. Changes the config to remote the allowed IP address. Checks that requests are passing.
3. Changes the config again to enable allowed IP address. Checks that requests are blocked.
'''


def run_test():
    response = php_server_get("/test", headers={"X-Forwarded-For": "2.20.116.1"})
    assert_response_code_is(response, 403)
    assert_response_header_contains(response, "Content-Type", "text")
    assert_response_body_contains(response, "Your ip (2.20.116.1) is blocked due to: not in allow lists!")

    response = php_server_get("/test", headers={"X-Forwarded-For": "2.17.116.2"})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Something")

    apply_config("change_config_remove_allow_list.json")
        
    response = php_server_get("/test", headers={"X-Forwarded-For": "2.20.116.1"})
    assert_response_code_is(response, 200)
    
    apply_config("start_config.json")
        
    response = php_server_get("/test", headers={"X-Forwarded-For": "2.20.116.1"})
    assert_response_code_is(response, 403)
    assert_response_header_contains(response, "Content-Type", "text")
    assert_response_body_contains(response, "Your ip (2.20.116.1) is blocked due to: not in allow lists!")

    response = php_server_get("/test", headers={"X-Forwarded-For": "2.17.116.2"})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Something")
    
    # Test that private IPs are always allowed even when allowlists are configured
    response = php_server_get("/test", headers={"X-Forwarded-For": "127.0.0.1"})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Something")
    
    
if __name__ == "__main__":
    load_test_args()
    run_test()
