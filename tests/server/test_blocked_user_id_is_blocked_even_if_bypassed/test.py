import requests
import time
import sys
from testlib import *

'''
1. Sets up a config with  "allowedIPAddresses": ["93.184.216.34"], and "blockedUserIds": ["12345"].
2. Sends a get request with the X-Forwarded-For header set to "93.184.216.34".
3. Verifies that the response code is 403; bypassed IPs should not bypass blocked user IDs.
'''

def run_test():
    response = php_server_get("/test", headers={"X-Forwarded-For": "93.184.216.34"})
    assert_response_code_is(response, 403)
    assert_response_header_contains(response, "Content-Type", "text")
    assert_response_body_contains(response, "You are blocked by Aikido Firewall!")

    
if __name__ == "__main__":
    load_test_args()
    run_test()
