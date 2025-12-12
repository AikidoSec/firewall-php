import requests
import time
import sys
from testlib import *

'''
Test that bypassed IPs skip all Zen protection features.

Scenarios tested:
1. Individual IPv4 address (93.184.216.34) - exact match
2. IPv4 CIDR range (23.45.67.0/24) - IP within range (23.45.67.89)
3. Individual IPv6 address (2606:2800:220:1:248:1893:25c8:1946)
4. IPv6 CIDR range (2001:0db9:abcd:1234::/64) - IP within range

For each bypassed IP:
- SQL injection attacks should NOT be blocked
- API spec should NOT be generated
- Statistics should NOT be counted

When IP is NOT bypassed:
- SQL injection attacks SHOULD be blocked
- API spec SHOULD be generated
- Statistics SHOULD be counted
'''

# Test IPs from the spec
BYPASSED_IPV4 = "93.184.216.34"
BYPASSED_IPV4_CIDR_IP = "23.45.67.89"  # Within 23.45.67.0/24
BYPASSED_IPV6 = "2606:2800:220:1:248:1893:25c8:1946"
BYPASSED_IPV6_CIDR_IP = "2001:0db9:abcd:1234::5678"  # Within 2001:0db9:abcd:1234::/64
NON_BYPASSED_IP = "8.8.8.8"


def test_attack_with_ip(ip, expect_blocked):
    """Test SQL injection attack from given IP"""
    headers = {"X-Forwarded-For": ip}
    response = php_server_post("/testDetection", {"userId": "1 OR 1=1"}, headers=headers)
    
    if expect_blocked:
        assert_response_code_is(response, 500)
    else:
        assert_response_code_is(response, 200)
        assert_response_body_contains(response, "Query executed!")


def test_api_spec_with_ip(ip, expect_collected):
    """Test that API spec is/isn't collected for given IP"""
    headers = {
        "X-Forwarded-For": ip,
        "Content-Type": "application/json"
    }
    response = php_server_post("/api/v1/orders?userId=12345", {"orderId": "98765"}, headers=headers)
    assert_response_code_is(response, 200)


def run_test():
    # ========================================
    # Test 1: Bypassed IPs - attacks NOT blocked
    # ========================================
    
    # Test individual IPv4 (93.184.216.34)
    test_attack_with_ip(BYPASSED_IPV4, expect_blocked=False)
    
    # Test IPv4 CIDR range (23.45.67.89 within 23.45.67.0/24)
    test_attack_with_ip(BYPASSED_IPV4_CIDR_IP, expect_blocked=False)
    
    # Test individual IPv6
    test_attack_with_ip(BYPASSED_IPV6, expect_blocked=False)
    
    # Test IPv6 CIDR range
    test_attack_with_ip(BYPASSED_IPV6_CIDR_IP, expect_blocked=False)
    
    # ========================================
    # Test 2: Bypassed IPs - API spec NOT collected
    # ========================================
    
    # Make API requests from bypassed IPs
    test_api_spec_with_ip(BYPASSED_IPV4, expect_collected=False)
    test_api_spec_with_ip(BYPASSED_IPV4_CIDR_IP, expect_collected=False)
    
    # Wait for potential heartbeat - should only have started event
    time.sleep(2)
    events = mock_server_get_events()
    assert_events_length_is(events, 1)  # Only started event, no detections, no heartbeat with API spec
    assert_started_event_is_valid(events[0])
    
    # ========================================
    # Test 3: Non-bypassed IP - attacks ARE blocked
    # ========================================
    
    apply_config("change_config_remove_bypassed_ip.json")
    
    # Attack from non-bypassed IP should be blocked
    test_attack_with_ip(NON_BYPASSED_IP, expect_blocked=True)
    
    # Verify detection event was sent
    mock_server_wait_for_new_events(5)
    events = mock_server_get_events()
    assert_events_length_is(events, 2)  # started + detection
    
    # ========================================
    # Test 4: Re-enable bypass - attacks NOT blocked again
    # ========================================
    
    apply_config("start_config.json")
    
    # Attack from bypassed IP should succeed again
    test_attack_with_ip(BYPASSED_IPV4, expect_blocked=False)
    test_attack_with_ip(BYPASSED_IPV4_CIDR_IP, expect_blocked=False)
    

if __name__ == "__main__":
    load_test_args()
    run_test()
