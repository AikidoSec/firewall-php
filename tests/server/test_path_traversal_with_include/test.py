import requests
import time
import sys
from testlib import *

'''
1. Sets up a simple config and env AIKIDO_BLOCK=1.
2. Sends an attack request to a route, that will cause sending a detection event.
3. Checks that the detection event was submitted and is valid.
'''

def check_path_traversal(response_code, response_body, event_id, expected_json):
    response = php_server_post("/testDetection", {"file": "../test_include.php"})
    assert_response_code_is(response, response_code)
    assert_response_body_contains(response, response_body)
    
    mock_server_wait_for_new_events(5)
    
    events = mock_server_get_events()
    assert_events_length_is(events, event_id + 1)
    assert_started_event_is_valid(events[0])
    assert_event_contains_subset_file(events[event_id], expected_json)

def run_test():
    check_path_traversal(500, "", 1, "expect_detection_blocked.json")
    
    apply_config("change_config_disable_blocking.json")
    check_path_traversal(200, "In test include", 2, "expect_detection_not_blocked.json")
        
if __name__ == "__main__":
    load_test_args()
    run_test()
