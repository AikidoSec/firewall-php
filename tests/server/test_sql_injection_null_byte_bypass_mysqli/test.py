import requests
import time
import sys
from testlib import *

'''
Test that SQL injection payloads containing embedded null bytes (\x00) are
still detected and blocked. Before the fix, null bytes caused truncation in
the C++/Go context pipeline, letting the detection logic see only the benign
prefix (e.g. "1") while MySQL executed the full payload.
'''

def check_null_byte_sqli_blocked(response_code, response_body, event_id, expected_json):
    response = php_server_post("/testDetection", {"userId": "1\x00 OR 1=1"})
    assert_response_code_is(response, response_code)
    assert_response_body_contains(response, response_body)

    mock_server_wait_for_new_events(5)

    events = mock_server_get_events()
    assert_events_length_is(events, event_id + 1)
    assert_started_event_is_valid(events[0])
    assert_event_contains_subset_file(events[event_id], expected_json)

def run_test():
    check_null_byte_sqli_blocked(500, "", 1, "expect_detection_blocked.json")

if __name__ == "__main__":
    load_test_args()
    run_test()
