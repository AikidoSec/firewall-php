import requests
import time
import sys
from testlib import *

'''
1. Sets up the allowed IP address config for route '/test'. Checks that requests are blocked.
2. Changes the config to remote the allowed IP address. Checks that requests are passing.
3. Changes the config again to enable allowed IP address. Checks that requests are blocked.
'''


def run_test():
    time.sleep(5)
    events = mock_server_get_events()
    assert_events_length_is(events, 0)

    for i in range(10):
        response = php_server_get("/somethingVerySpecific")
        assert_response_code_is(response, 200)
        assert_response_body_contains(response, "Something")

    mock_server_wait_for_new_events(60)
    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_started_event_is_valid(events[0])

    token = mock_server_get_token()
    assert token == "my-test-token", "Token is %s, should be my-test-token" % token
    
    
if __name__ == "__main__":
    load_test_args()
    run_test()
