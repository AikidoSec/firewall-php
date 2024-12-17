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
    response = php_server_get("/test")
    assert_response_code_is(response, 403)
    assert_response_header_contains(response, "Content-Type", "text")
    assert_response_body_contains(response, "Your IP (5.8.19.22) is blocked due to: geo restrictions!")

    apply_config("change_config_remove_geo_blocked_ips.json")
        
    response = php_server_get("/test")
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Something")
    
    apply_config("start_config.json")
        
    response = php_server_get("/test")
    assert_response_code_is(response, 403)
    assert_response_header_contains(response, "Content-Type", "text")
    assert_response_body_contains(response, "Your IP (5.8.19.22) is blocked due to: geo restrictions!")
    
    
if __name__ == "__main__":
    load_test_args()
    run_test()
