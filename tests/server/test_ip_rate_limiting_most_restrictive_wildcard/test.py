import requests
import time
import sys
from testlib import *

'''
1. Sets up the rate limiting config to 5 requests / minute for route '/'.
2. Sends 5 requests to '/'. Checks that those requests are not blocked.
3. Send another more 5 request to '/'. Checks that they all are rate limited.
4. Sends 100 requests to another route '/tests'. Checks that those requests are not blocked.
5. Tests wildcard route specificity with /api/* vs /api/*/auth/* routes.
'''

def run_test():
    for _ in range(5):
        response = php_server_get("/test" + generate_random_string(3))
        assert_response_code_is(response, 200)
        
    time.sleep(10)
        
    for _ in range(5):
        response = php_server_get("/test" + generate_random_string(3))
        assert_response_code_is(response, 429)
        assert_response_header_contains(response, "Content-Type", "text")
        assert_response_body_contains(response, "Rate limit exceeded")

    # /api/v2/auth/login should match both:
    # - /api/* (50 requests/minute)
    # - /api/*/auth/* (3 requests/minute)
    # The more specific /api/*/auth/* should take precedence
    for i in range(3):
        response = php_server_get("/api/v2/auth/login")
        assert_response_code_is(response, 200)

    response = php_server_get("/api/v2/auth/login")
    assert_response_code_is(response, 429)
    assert_response_header_contains(response, "Content-Type", "text")
    assert_response_body_contains(response, "Rate limit exceeded")

if __name__ == "__main__":
    load_test_args()
    run_test()
