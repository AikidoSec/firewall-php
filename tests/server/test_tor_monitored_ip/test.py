import requests
import time
import sys
from testlib import *

'''
1. Sets up the monitored IP addresses. Checks that requests are blocked.
2. Changes the config to remote the allowed IP address. Checks that requests are passing.
3. Changes the config again to enable allowed IP address. Checks that requests are blocked.
'''


def run_test():
    response = php_server_get("/test", headers={"X-Forwarded-For": "192.42.116.197"})
    assert_response_code_is(response, 200)
    assert_response_header_contains(response, "Content-Type", "text")
    assert_response_body_contains(response, "Something")
    
    mock_server_wait_for_new_events(65)

    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_started_event_is_valid(events[0])
    
    assert_event_contains_subset_file(events[1], "expect_ip_monitored.json")
    
    
if __name__ == "__main__":
    load_test_args()
    run_test()
