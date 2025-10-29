import requests
import time
import sys
from testlib import *

'''
1. Sets up receiveAnyStats to false.
2. Checks that the 'started' event is valid.
3. After 1 minute, checks that the heartbeat event was sent.
'''

def run_test():
    php_server_get("/")
    events = mock_server_get_events()
    assert_events_length_is(events, 1)
    assert_started_event_is_valid(events[0])
    
    mock_server_wait_for_new_events(70)
    
    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_event_contains_subset("__root", events[1], {"type": "heartbeat" })
    
if __name__ == "__main__":
    load_test_args()
    run_test()
