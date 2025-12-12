import requests
import time
import sys
from testlib import *

'''
1. Non-allowed IP is not blocked if bypassed.
'''


def run_test():        
    response = php_server_get("/somethingVerySpecific", headers={"X-Forwarded-For": "185.245.255.211"})
    assert_response_code_is(response, 200)
    assert_response_header_contains(response, "Content-Type", "text")
    assert_response_body_contains(response, "Something")
    
    
if __name__ == "__main__":
    load_test_args()
    run_test()
