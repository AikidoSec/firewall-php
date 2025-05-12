import requests
import time
import sys
from testlib import *

'''
1. Sets up the bypassed IP address config for route '/test'. Rate limiting is set to 10 req / min. Checks that requests are not rate limited blocked.
2. Changes the config to remove the bypassed IP address. Checks that requests are rate limiting.
3. Changes the config again to enable the bypassed IP address. Checks that requests are not rate limited blocked.
'''


def run_test():
    for _ in range(100):
        response = php_server_get("/test")
        assert_response_code_is(response, 200)
        
    apply_config("change_config_remove_bypassed_ip.json")

    for i in range(100):
        response = php_server_get("/test")
        if i < 10:
            assert_response_code_is(response, 200)
        else:
            assert_response_code_is(response, 429)
            assert_response_header_contains(response, "Content-Type", "text")
            assert_response_body_contains(response, "is blocked due to: configured rate limit exceeded by current ip")

    apply_config("start_config.json")
    
    for _ in range(100):
        response = php_server_get("/test")
        assert_response_code_is(response, 200)
    
    
if __name__ == "__main__":
    load_test_args()
    run_test()
