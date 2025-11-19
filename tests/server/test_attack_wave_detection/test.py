import requests
import time
import sys
from testlib import *

'''
1. Send 15 requests with web-scanner paths or queries to the server.
2. Wait for the detection event and validate it.
3. Send 15 requests with the same IP, no new event should be sent (only one event should be sent per IP in 20 minutes window).
4. Send 15 requests with a new IP, a new event should be sent (different IP).
5. Repeat steps 3 and 4 for a new IP.
'''


paths = [
    "/path/?test=' or '1'='1",
    "/path/?test=1: SELECT * FROM users WHERE '1'='1'",
    "/path/?test=', information_schema.tables",
    "/path/?test=1' sleep(5)",
    "/path/?test=WAITFOR DELAY 1",
    "/path/?test=../etc/passwd",
     "/etc/passwd"
]


def get_random_path():
    return random.choice(paths)

def run_test():
    for i in range(15):
        _ = php_server_get(get_random_path(), headers={"X-Forwarded-For": "5.8.19.22"})

    mock_server_wait_for_new_events(5)
    
    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_started_event_is_valid(events[0])
    assert_event_contains_subset_file(events[1], "expect_wave_detection.json")

    time.sleep(70)

    for i in range(15):
        _ = php_server_get(get_random_path(), headers={"X-Forwarded-For": "5.8.19.22"})
    mock_server_wait_for_new_events(5)
    
    events = mock_server_get_events()
    assert_events_length_is(events, 2) # no new event should be sent (same IP)

    for i in range(15):
        _ = php_server_get(get_random_path(), headers={"X-Forwarded-For": "5.8.19.23"})
    mock_server_wait_for_new_events(5)
    
    events = mock_server_get_events()
    assert_events_length_is(events, 3) # new event should be sent (new IP)
    assert_event_contains_subset_file(events[2], "expect_wave_detection_2.json")

    for i in range(15):
        _ = php_server_get(get_random_path(), headers={"X-Forwarded-For": "5.8.19.23"})
    mock_server_wait_for_new_events(5)
    
    events = mock_server_get_events()
    assert_events_length_is(events, 3) # no new event should be sent (same IP)



if __name__ == "__main__":
    load_test_args()
    run_test()
