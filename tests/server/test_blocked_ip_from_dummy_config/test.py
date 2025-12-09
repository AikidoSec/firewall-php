import requests
import time
import sys
from testlib import *

'''
Test that verifies an IP from the dummy start_config.json is actually blocked.
1. Uses a hardcoded IP that is in the blockedIPAddresses in start_config.json
2. Makes a request with that IP in X-Forwarded-For header
3. Verifies the request is blocked (403 response)
4. Verifies the response contains the expected blocking message
'''


def run_test():
    test_ip = "76.54.32.21"
    description = "geo restrictions"
    print(f"Testing with IP: {test_ip} (should be blocked due to: {description})")

    # Make a request with the blocked IP
    response = php_server_get("/", headers={"X-Forwarded-For": test_ip})

    # Verify the request is blocked
    assert_response_code_is(response, 403)
    assert_response_header_contains(response, "Content-Type", "text")

    # Verify the response contains the blocking message
    expected_message = f"Your ip ({test_ip}) is blocked due to: {description}!"
    assert_response_body_contains(response, expected_message)


if __name__ == "__main__":
    load_test_args()
    run_test()
