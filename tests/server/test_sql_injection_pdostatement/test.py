import requests
import time
import sys
from testlib import *

'''
1. Sets up a simple config and env AIKIDO_BLOCK=1.
2. Sends an attack request to a route, that will cause sending a detection event.
3. Checks that the detection event was submitted and is valid.
'''

def check_sql_injection(response_code, response_body, event_id, expected_json):
    response = php_server_get("/testDetection?name=Tom&age=3%27%29%3B%20DROP%20TABLE%20cats%3B%20--")
    assert_response_code_is(response, response_code)
    assert_response_body_contains(response, response_body)
    
    mock_server_wait_for_new_events(5)
    
    events = mock_server_get_events()
    assert_events_length_is(events, event_id + 1)
    assert_started_event_is_valid(events[0])
    assert_event_contains_subset_file(events[event_id], expected_json)

def run_test():
    check_sql_injection(500, "", 1, "expect_detection_blocked_pdostatement.json")
    
    apply_config("change_config_disable_blocking.json")
    check_sql_injection(200, "Query executed!", 2, "expect_detection_not_blocked_pdostatement.json")
    
    apply_config("start_config.json")
    check_sql_injection(500, "", 3, "expect_detection_blocked_pdostatement.json")
    
if __name__ == "__main__":
    load_test_args()
    run_test()
