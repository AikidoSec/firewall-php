import requests
import time
import sys
from testlib import *

def run_test():
    ips = ["80.80.80.80", "90.90.90.90", "123.123.123.123"]
    for ip in ips:
        for _ in range(10):
            response = php_server_post("/", { "ip": ip })
            assert_response_code_is(response, 200)
            assert_response_header_contains(response, "Content-Type", "text")
            assert_response_body_contains(response, "Request successful")
            
    for ip in ips:
        php_server_post("/", { "ip": ip })
        
    time.sleep(5)
        
    for ip in ips:    
        for _ in range(10):
            response = php_server_post("/", { "ip": ip })
            assert_response_code_is(response, 429)
            assert_response_header_contains(response, "Content-Type", "text")
            assert_response_body_contains(response, "Rate limit exceeded")
                
    
if __name__ == "__main__":
    load_test_args()
    run_test()
