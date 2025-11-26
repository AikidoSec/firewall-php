import requests
import time
import sys
from testlib import *

'''
Tests the outbound domain blocking feature:
1. Tests that explicitly blocked domains are always blocked
2. Tests that allowed domains can be accessed when blockNewOutgoingRequests is true
3. Tests that new/unknown domains are blocked when blockNewOutgoingRequests is true
4. Tests that new domains are allowed when blockNewOutgoingRequests is false
5. Tests that explicitly blocked domains are still blocked when blockNewOutgoingRequests is false
6. Tests that detection mode (block: false) doesn't block
'''

def test_explicitly_blocked_domain():
    """Test that explicitly blocked domains are always blocked"""
    response = php_server_post("/testDetection", {"url": "http://evil.example.com"})
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
    test_allowed_domain_with_block_new()
    test_new_domain_blocked_when_flag_enabled()
    test_new_domain_allowed_when_flag_disabled()
    test_blocked_domain_still_blocked_when_flag_disabled()
    test_detection_mode()
    test_case_insensitive_matching()
    
if __name__ == "__main__":
    load_test_args()
    run_test()

