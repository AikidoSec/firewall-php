import requests
import time
import sys
from testlib import *

'''
Test that bypassed IPs skip rate limiting.

1. Sets up bypassed IP addresses (IPv4 + CIDR range from spec). Rate limiting is set to 10 req / min.
2. Makes 100 requests from bypassed IP - should NOT be rate limited.
3. Changes config to remove bypassed IPs.
4. Makes requests from same IP - should be rate limited after 10 requests.
5. Re-enables bypassed IPs - should NOT be rate limited again.
'''

BYPASSED_IP = "93.184.216.34"
BYPASSED_IP_CIDR = "23.45.67.89"  # Within 23.45.67.0/24
NON_BYPASSED_IP = "8.8.8.8"


def run_test():
    # Test 1: Bypassed IP - no rate limiting
    for _ in range(100):
        response = php_server_get("/test", headers={"X-Forwarded-For": BYPASSED_IP})
        assert_response_code_is(response, 200)
    
    # Test 2: Bypassed IP from CIDR range - no rate limiting
    for _ in range(100):
        response = php_server_get("/test", headers={"X-Forwarded-For": BYPASSED_IP_CIDR})
        assert_response_code_is(response, 200)
        
    # Test 3: Remove bypass - rate limiting kicks in
    apply_config("change_config_remove_bypassed_ip.json")

    for i in range(100):
        response = php_server_get("/test", headers={"X-Forwarded-For": NON_BYPASSED_IP})
        if i < 10:
            assert_response_code_is(response, 200)
        else:
            assert_response_code_is(response, 429)
            assert_response_header_contains(response, "Content-Type", "text")
            assert_response_body_contains(response, "Rate limit exceeded")

    # Test 4: Re-enable bypass - no rate limiting again
    apply_config("start_config.json")
    
    for _ in range(100):
        response = php_server_get("/test", headers={"X-Forwarded-For": BYPASSED_IP})
        assert_response_code_is(response, 200)
    
    
if __name__ == "__main__":
    load_test_args()
    run_test()
