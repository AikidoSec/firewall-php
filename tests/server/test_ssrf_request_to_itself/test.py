import requests
import time
import sys
from testlib import *

'''
Test for SSRF "request to itself" functionality.
This test verifies that:
1. Requests to the same server (localhost) are NOT blocked (no false positives)
2. Requests to different hosts are still properly blocked
3. The HTTP/HTTPS special case is handled correctly
'''

def check_request_to_itself(url, response_code, response_body):
    """
    Verify that a request to itself is NOT blocked and does NOT generate a detection event
    """
    
    response = php_server_post("/testDetection", {"url": url})
    assert_response_code_is(response, response_code)
    assert_response_body_contains(response, response_body)
    
def run_test():
    php_port = get_php_port()
    
    print(f"\n=== Testing SSRF Request-to-Itself Prevention ===")
    print(f"PHP Server Port: {php_port}")
  
    # Test 1: Request to localhost (same server) should NOT be blocked
    # This simulates a server making a request to itself on the same port
    print("\n[Test 1] Request to itself (localhost) should not be blocked...")
    check_request_to_itself(
        f"http://localhost:{php_port}/test",
        200,
        "Got URL content!"
    )
    
    # Test 2: [HTTPS] Request to localhost (same server) should NOT be blocked
    print("\n[Test 2] Request to itself (localhost - HTTPS) should not be blocked...")
    check_request_to_itself(
        f"https//localhost:{php_port}/test",
        200,
        "Got URL content!"
    )

    print("\n[Test 3] Request to itself, but with a different port should be blocked...")
    check_request_to_itself(
        f"http://localhost:{php_port+100}/test",
        500,
        "Aikido firewall has blocked a server-side request forgery"
    )
    
    print("\n=== All tests passed! ===\n")
    
if __name__ == "__main__":
    load_test_args()
    run_test()

