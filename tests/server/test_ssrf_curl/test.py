import requests
import time
import sys
from testlib import *

'''
1. Sets up a simple config and env AIKIDO_BLOCK=1.
2. Sends an attack request to a route, that will cause sending a detection event.
3. Checks that the detection event was submitted and is valid.
'''

def check_ssrf(url, response_code, response_body, event_id, expected_json):
    response = php_server_post("/testDetection", {"url": url})
    assert_response_code_is(response, response_code)
    assert_response_body_contains(response, response_body)
    
    mock_server_wait_for_new_events(5)
    
    events = mock_server_get_events()
    assert_events_length_is(events, event_id + 1)
    assert_started_event_is_valid(events[0])
    assert_event_contains_subset_file(events[event_id], expected_json)


def check_false_positive_ssrf(url, response_code):
    response = php_server_post("/testDetection", {"url": url})
    assert_response_code_is(response, response_code)

def run_test():
    check_ssrf("http://127.0.0.1:8081", 500, "", 1, "expect_detection_blocked.json")
    
    add_to_hosts_file("app.example.local", "127.0.0.1")
    
    apply_config("change_config_disable_blocking.json")
    check_ssrf("http://127.0.0.1:8081", 200, "Got URL content!", 2, "expect_detection_not_blocked.json")
    
    apply_config("start_config.json")
    
    check_ssrf(f"http://app.example.local:{get_mock_port()}/tests/simple", 500, "", 3, "expect_detection_blocked_resolved_ip.json")
    
    # http://ssrf-redirects.testssandbox.com/ssrf-test-domain -> http://local.aikido.io/test
    # add a local.aikido.io to the hosts file and check that the request is not blocked (public IP address)
    add_to_hosts_file("local.aikido.io", "8.8.8.8")
    check_false_positive_ssrf(f"http://ssrf-redirects.testssandbox.com/ssrf-test-domain", 200)

if __name__ == "__main__":
    load_test_args()
    run_test()
