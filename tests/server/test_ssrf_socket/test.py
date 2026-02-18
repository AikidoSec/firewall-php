import requests
import time
import sys
from testlib import *

'''
Test socket functions (socket_connect, fsockopen, stream_socket_client) for SSRF detection:
1. Tests that socket_connect is detected and blocked when accessing localhost
2. Tests that fsockopen is detected and blocked when accessing localhost
3. Tests that stream_socket_client is detected and blocked when accessing localhost
4. Tests that blocking can be disabled via config
5. Tests that resolved IPs are properly detected
'''

def check_ssrf_socket(url, method, response_code, response_body, event_id, expected_json):
    """Test SSRF detection for socket functions"""
    payload = {"url": url}
    if method:
        payload["method"] = method
    
    response = php_server_post("/testDetection", payload)
    assert_response_code_is(response, response_code)
    if response_body:
        assert_response_body_contains(response, response_body)
    
    if event_id >= 0:
        mock_server_wait_for_new_events(5)
        
        events = mock_server_get_events()
        assert_events_length_is(events, event_id + 1)
        assert_started_event_is_valid(events[0])
        assert_event_contains_subset_file(events[event_id], expected_json)

def run_test():
    # Test socket_connect - should be blocked
    check_ssrf_socket("http://127.0.0.1:8081", "socket_connect", 500, "", 1, "expect_detection_blocked.json")
    
    # Test fsockopen - should be blocked
    check_ssrf_socket("http://127.0.0.1:8081", "fsockopen", 500, "", 2, "expect_detection_blocked_fsockopen.json")
    
    # Test stream_socket_client - should be blocked
    check_ssrf_socket("http://127.0.0.1:8081", "stream_socket_client", 500, "", 3, "expect_detection_blocked_stream_socket_client.json")
    
    # Add hostname mapping for resolved IP test
    add_to_hosts_file("app1.example.local", "127.0.0.1")
    
    # Disable blocking - should detect but not block
    apply_config("change_config_disable_blocking.json")
    check_ssrf_socket("http://127.0.0.1:8081", "socket_connect", 200, "Socket connect completed!", 4, "expect_detection_not_blocked.json")
    
    # Re-enable blocking
    apply_config("start_config.json")
    
    # Test with resolved IP (hostname that resolves to localhost)
    check_ssrf_socket(f"http://app1.example.local:{get_mock_port()}/tests/simple", "socket_connect", 500, "", 5, "expect_detection_blocked_resolved_ip.json")
    
    print("\n=== All socket SSRF tests passed! ===\n")

if __name__ == "__main__":
    load_test_args()
    run_test()

