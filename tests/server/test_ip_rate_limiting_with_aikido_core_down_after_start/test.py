import requests
import time
import sys
from testlib import *

'''
1. Bring the mock server down (similate aikido core going down after the started event so we can pull the rate limiting config).
2. Sets up the rate limiting config to 5 requests / minute for route '/'.
3. Sends 5 requests to '/'. Checks that those requests are not blocked.
4. Send another more 5 request to '/'. Checks that they all are rate limited.
5. Sends 100 requests to another route '/tests'. Checks that those requests are not blocked.
'''

def run_test():
    response = php_server_get("/test")
    time.sleep(5)

    mock_server_down()

    for _ in range(5):
        response = php_server_get("/")
        assert_response_code_is(response, 200)

    time.sleep(10)

    for _ in range(5):
        response = php_server_get("/")

    for _ in range(5):
        response = php_server_get("/")
        assert_response_code_is(response, 429)
        assert_response_header_contains(response, "Content-Type", "text")
        assert_response_body_contains(response, "Rate limit exceeded")

    for _ in range(100):
        response = php_server_get("/test")
        assert_response_code_is(response, 200)
        assert_response_body_contains(response, "Request successful")
    
if __name__ == "__main__":
    load_test_args()
    run_test()
