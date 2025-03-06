import requests
import time
import sys
from testlib import *

'''
1. Sets up a simple config.
2. Makes a request and checks that 4 detections are emitted for that route.
3. Changes the config to force protection off for that route and checks that no other detections are emitted.
4. Puts the initial config back, makes the request and checks that the 4 detections are still emitted.
'''

def assert_detection_events_are_valid(events):
    for i in range(4):
        assert_detection_event_is_valid(events[i])

def run_test():
    data = {
        "command": "`whoami`", 
        "folder": "../../../..",
        "url": "http://127.0.0.1:8081",
        "userId": "1 OR 1=1"
    }
    assert_response_code_is(php_server_post("/", data=data), 200)
    time.sleep(10)
    events = mock_server_get_events()
    assert_events_length_is(events, 5)
    assert_started_event_is_valid(events[0])
    assert_detection_events_are_valid(events[1:5])
        
    apply_config("change_config_set_force_protection_off.json")
    time.sleep(120)
    
    assert_response_code_is(php_server_post("/", data=data), 200)
    time.sleep(10)
    events = mock_server_get_events()
    assert_events_length_is(events, 5)
    
    apply_config("start_config.json")
    time.sleep(120)
    
    assert_response_code_is(php_server_post("/", data=data), 200)
    time.sleep(10)
    events = mock_server_get_events()
    assert_events_length_is(events, 9)
    assert_detection_events_are_valid(events[5:])
        
    
if __name__ == "__main__":
    load_test_args()
    run_test()
