import requests
import time
import sys
from testlib import *

'''
1. Sets up the rate limiting config to 30 requests / 10 minutes.
2. Sends 10 requests once every minute, for 3 minutes. Checks that those requests are not blocked.
3. Send another more 10 request. Checks that they all are rate limited.
'''

def run_test():
    for i in range(30):
        response = php_server_get("/test")
        assert_response_code_is(response, 200)
        assert_response_body_contains(response, "Request successful")
        
        if i != 0 and i % 10 == 0:
            time.sleep(60)
        
    for _ in range(10):
        response = php_server_get("/test")
        assert_response_code_is(response, 429)
        assert_response_header_contains(response, "Content-Type", "text")
        assert_response_body_contains(response, "Rate limit exceeded")
    
    
if __name__ == "__main__":
    load_test_args()
    run_test()
