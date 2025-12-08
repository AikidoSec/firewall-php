import requests
import time
import sys
from testlib import *

'''
Tests the outbound domain blocking feature:
1. Tests that explicitly blocked domains are always blocked
2. Tests that bypassed IPs (allowedIPAddresses) can access any domain including blocked ones
3. Tests that non-bypassed IPs are blocked when accessing new domains with blockNewOutgoingRequests enabled
4. Tests that forceProtectionOff does not affect outbound domain blocking
5. Tests that allowed domains can be accessed when blockNewOutgoingRequests is true
6. Tests that new/unknown domains are blocked when blockNewOutgoingRequests is true
7. Tests that new domains are allowed when blockNewOutgoingRequests is false
8. Tests that explicitly blocked domains are still blocked when blockNewOutgoingRequests is false
9. Tests that detection mode (block: false) doesn't block
10. Tests case-insensitive hostname matching
'''

def test_explicitly_blocked_domain():
    """Test that explicitly blocked domains are always blocked"""
    response = php_server_post("/testDetection", {"url": "http://evil.example.com/test"})
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "Aikido firewall has blocked an outbound connection")

    response = php_server_post("/testDetection", {"url": "http://evil.example.com/test"}, headers={"X-Forwarded-For": "1.2.3.4"})
    assert_response_code_is(response, 200)

    response = php_server_post("/testDetection", {"url": "http://random.example.com"}, headers={"X-Forwarded-For": "1.2.3.4"})
    assert_response_code_is(response, 200)

    response = php_server_post("/testDetection", {"url": "http://random2.example.com/test"}, headers={"X-Forwarded-For": "1.2.3.5"})
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "Aikido firewall has blocked an outbound connection")
    
    mock_server_wait_for_new_events(70)
    
    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_started_event_is_valid(events[0])
    
    assert_event_contains_subset_file(events[1], "blocked_domains_in_heartbeat.json")

def test_force_protection_off():
    """Test that force protection off does not affect outbound domain blocking"""
    response = php_server_post("/testDetection2", {"url": "http://evil.example.com/test"})
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "Aikido firewall has blocked an outbound connection")


def test_allowed_domain_with_block_new():
    """Test that allowed domains can be accessed when blockNewOutgoingRequests is true"""
    response = php_server_post("/testDetection", {"url": "http://safe.example.com"})
    assert_response_code_is(response, 200)
   
def test_new_domain_blocked_when_flag_enabled():
    """Test that new/unknown domains are blocked when blockNewOutgoingRequests is true"""
    response = php_server_post("/testDetection", {"url": "http://unknown.example.com"})
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "Aikido firewall has blocked an outbound connection")

def test_new_domain_allowed_when_flag_disabled():
    """Test that new domains are allowed when blockNewOutgoingRequests is false"""
    apply_config("config_disable_block_new.json")
    
    response = php_server_post("/testDetection", {"url": "http://another-unknown.example.com"})
    assert_response_code_is(response, 200)

def test_blocked_domain_still_blocked_when_flag_disabled():
    """Test that explicitly blocked domains are still blocked when blockNewOutgoingRequests is false"""
    response = php_server_post("/testDetection", {"url": "http://evil.example.com"})
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "Aikido firewall has blocked an outbound connection")
   
def test_detection_mode():
    """Test that detection mode (block: false) detects but doesn't block"""
    apply_config("config_no_blocking.json")
    
    response = php_server_post("/testDetection", {"url": "http://evil.example.com"})
    assert_response_code_is(response, 200)

def test_case_insensitive_matching():
    """Test that hostname matching is case-insensitive"""
    apply_config("start_config.json")
    
    # Test with uppercase hostname
    response = php_server_post("/testDetection", {"url": "http://EVIL.EXAMPLE.COM"})
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "Aikido firewall has blocked an outbound connection")
   
    # Test with mixed case
    response = php_server_post("/testDetection", {"url": "http://Evil.Example.Com"})
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "Aikido firewall has blocked an outbound connection")
  
    
def run_test():
    test_explicitly_blocked_domain()
    test_force_protection_off()
    test_allowed_domain_with_block_new()
    test_new_domain_blocked_when_flag_enabled()
    test_new_domain_allowed_when_flag_disabled()
    test_blocked_domain_still_blocked_when_flag_disabled()
    test_detection_mode()
    test_case_insensitive_matching()
    
if __name__ == "__main__":
    load_test_args()
    run_test()

