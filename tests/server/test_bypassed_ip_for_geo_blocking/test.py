import requests
import time
import sys
from testlib import *

'''
Test that bypassed IPs skip geo blocking.

1. Request from geo-blocked IP that is also in bypass list - should NOT be blocked
2. Remove IP from bypass list - request should be blocked for geo restrictions
3. Re-add IP to bypass list - should NOT be blocked again

Uses IP 5.8.19.22 which is in the geo-blocked range (5.8.19.0/24) but also in bypass list.
'''

# IP that is geo-blocked but also bypassed
GEO_BLOCKED_BUT_BYPASSED_IP = "5.8.19.22"


def run_test():
    # Test 1: IP is bypassed - geo blocking is skipped
    response = php_server_get("/test", headers={"X-Forwarded-For": GEO_BLOCKED_BUT_BYPASSED_IP})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Something")

    # Test 2: Remove IP from bypass list - geo blocking takes effect
    apply_config("change_config_remove_bypassed_ip.json")
        
    response = php_server_get("/test", headers={"X-Forwarded-For": GEO_BLOCKED_BUT_BYPASSED_IP})
    assert_response_code_is(response, 403)
    assert_response_header_contains(response, "Content-Type", "text")
    assert_response_body_contains(response, "is blocked due to: geo restrictions!")
    
    # Test 3: Re-add IP to bypass list - geo blocking is skipped again
    apply_config("start_config.json")
        
    response = php_server_get("/test", headers={"X-Forwarded-For": GEO_BLOCKED_BUT_BYPASSED_IP})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Something")
    
    
if __name__ == "__main__":
    load_test_args()
    run_test()
