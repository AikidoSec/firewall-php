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
11. Tests Punycode/IDN bypass prevention (Unicode domain blocked, Punycode request)
12. Tests reverse Punycode bypass prevention (Punycode domain blocked, Unicode request)
13. Tests URL percent-encoding bypass prevention
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
    
    response = php_server_post("/testDetection", {"url": "http://another-unknown.example.com"})
    assert_response_code_is(response, 200)

def test_blocked_domain_still_blocked_when_flag_disabled():
    """Test that explicitly blocked domains are still blocked when blockNewOutgoingRequests is false"""
    response = php_server_post("/testDetection", {"url": "http://evil.example.com"})
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "Aikido firewall has blocked an outbound connection")
   
def test_detection_mode():
    """Test that detection mode (block: false) detects but doesn't block"""
    
    response = php_server_post("/testDetection", {"url": "http://evil.example.com"})
    assert_response_code_is(response, 200)

def test_case_insensitive_matching():
    """Test that hostname matching is case-insensitive"""
    
    # Test with uppercase hostname
    response = php_server_post("/testDetection", {"url": "http://EVIL.EXAMPLE.COM"})
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "Aikido firewall has blocked an outbound connection")
   
    # Test with mixed case
    response = php_server_post("/testDetection", {"url": "http://Evil.Example.Com"})
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "Aikido firewall has blocked an outbound connection")

def test_punycode_bypass_unicode_blocked_punycode_request():
    """Test that Punycode requests are blocked when Unicode domain is in blocklist.
    Config has 'böse.example.com' blocked, attacker tries 'xn--bse-sna.example.com' (Punycode)"""
    
    # böse.example.com is blocked in config as Unicode
    # xn--bse-sna is the Punycode encoding of "böse"
    response = php_server_post("/testDetection", {"url": "http://xn--bse-sna.example.com"})
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "Aikido firewall has blocked an outbound connection")

def test_punycode_bypass_punycode_blocked_unicode_request():
    """Test that Unicode requests are blocked when Punycode domain is in blocklist.
    Config has 'xn--mnchen-3ya.example.com' blocked, attacker tries 'münchen.example.com' (Unicode)"""
    
    # xn--mnchen-3ya.example.com is blocked in config as Punycode
    # münchen is the Unicode form
    response = php_server_post("/testDetection", {"url": "http://münchen.example.com"})
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "Aikido firewall has blocked an outbound connection")

def test_punycode_allowed_domain():
    """Test that allowed IDN domains work with both Unicode and Punycode forms"""
    
    # münchen-allowed.example.com is allowed in config as Unicode
    # Should work with Unicode form
    response = php_server_post("/testDetection", {"url": "http://münchen-allowed.example.com"})
    assert_response_code_is(response, 200)
    
    # Should also work with Punycode form (xn--mnchen-allowed-gsb.example.com)
    response = php_server_post("/testDetection", {"url": "http://xn--mnchen-allowed-gsb.example.com"})
    assert_response_code_is(response, 200)

def test_url_percent_encoding_bypass():
    """Test that URL percent-encoded hostnames are properly normalized.
    Attackers might try to use %C3%B6 instead of ö to bypass blocking."""
    
    # böse.example.com is blocked - try with percent-encoded ö (%C3%B6)
    # b%C3%B6se.example.com should be normalized to böse.example.com
    response = php_server_post("/testDetection", {"url": "http://b%C3%B6se.example.com"})
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "Aikido firewall has blocked an outbound connection")
  
    
def run_test():
    test_explicitly_blocked_domain()
    test_force_protection_off()
    test_allowed_domain_with_block_new()
    test_new_domain_blocked_when_flag_enabled()
    test_case_insensitive_matching()
    test_punycode_bypass_unicode_blocked_punycode_request()
    test_punycode_bypass_punycode_blocked_unicode_request()
    test_punycode_allowed_domain()
    test_url_percent_encoding_bypass()
    apply_config("config_disable_block_new.json")
    test_new_domain_allowed_when_flag_disabled()
    test_blocked_domain_still_blocked_when_flag_disabled()
    apply_config("config_no_blocking.json")
    test_detection_mode()

    
if __name__ == "__main__":
    load_test_args()
    run_test()

