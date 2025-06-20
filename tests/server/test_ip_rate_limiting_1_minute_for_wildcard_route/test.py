import requests
import time
import sys
from testlib import *

'''
1. Sets up the rate limiting config to 5 requests / minute for route '/'.
2. Sends 5 requests to '/'. Checks that those requests are not blocked.
3. Send another more 5 request to '/'. Checks that they all are rate limited.
4. Sends 100 requests to another route '/tests'. Checks that those requests are not blocked.
5. Sends requests to '/login' route with mixed HTTP method to check rate limiting.
'''

def run_test():
    for _ in range(5):
        response = php_server_get("/myroute/" + generate_random_string(10))
        assert_response_code_is(response, 200)
        
    time.sleep(10)
    
    for _ in range(10):
        response = php_server_get("/myroute/" + generate_random_string(10))
        assert_response_code_is(response, 429)
        assert_response_header_contains(response, "Content-Type", "text")
        assert_response_body_contains(response, "Rate limit exceeded")
    
    for _ in range(100):
        response = php_server_get("/test")
        assert_response_code_is(response, 200)
        assert_response_body_contains(response, "Request successful")

    for _ in range(2):
        response = php_server_get("/login")
        assert_response_code_is(response, 200)

    response = php_server_post("/login", data={})
    assert_response_code_is(response, 200)

    response = php_server_get("/login")
    assert_response_code_is(response, 429)
    assert_response_header_contains(response, "Content-Type", "text")
    assert_response_body_contains(response, "Rate limit exceeded")

    response = php_server_post("/login", data={})
    assert_response_code_is(response, 429)
    assert_response_header_contains(response, "Content-Type", "text")
    assert_response_body_contains(response, "Rate limit exceeded")

    mock_server_wait_for_new_events(60)
    
    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_started_event_is_valid(events[0])
    assert_event_contains_subset_file(events[1], "expect_rate_limiting.json")
    
if __name__ == "__main__":
    load_test_args()
    run_test()
