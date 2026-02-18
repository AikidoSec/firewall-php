import requests
import time
import sys
from testlib import *

'''
1. Sets up a simple config.
2. Sends one request that will trigger multiple curl reuqests from php.
3. Waits for the heartbeat event and validates it.
'''

def run_test():
    response = php_server_get("/")
    assert_response_code_is(response, 200)
    
    mock_server_wait_for_new_events(70)
    
    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_started_event_is_valid(events[0])
    
    with open("expect_domains.json", 'r') as file:
        expected = json.load(file)
    
    if mock_server_get_platform_name() == "frankenphp":
        expected["hostnames"] = [h for h in expected["hostnames"] if h.get("hostname") != "127.0.0.1"]
    
    assert_event_contains_subset("__root", events[1], expected)
    
if __name__ == "__main__":
    load_test_args()
    run_test()
