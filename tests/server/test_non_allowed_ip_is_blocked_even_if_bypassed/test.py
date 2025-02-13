import requests
import time
import sys
from testlib import *

'''
1. Sets up the allowed IP address config and a bypassed IP for route '/somethingVerySpecific'. Checks that requests are NOT blocked.
2. Changes the config to remote the bypassed IP address. Checks that requests are blocked.
3. Changes the config again to enable bypassed IP address. Checks that requests are passing.
'''


def run_test():        
    response = php_server_get("/somethingVerySpecific")
    assert_response_code_is(response, 403)
    assert_response_header_contains(response, "Content-Type", "text")
    assert_response_body_contains(response, "is blocked due to: not allowed by config to access this endpoint!")
    
    
if __name__ == "__main__":
    load_test_args()
    run_test()
