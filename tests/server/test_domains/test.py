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
    assert_event_contains_subset_file(events[1], "expect_domains.json")
    
if __name__ == "__main__":
    load_test_args()
    run_test()
