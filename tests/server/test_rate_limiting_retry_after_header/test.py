import requests
import time
import sys
from testlib import *

'''
1. Sets up rate limiting config to 3 requests / 5 minutes for route '/'.
2. Sends 3 requests to '/'. Checks that those requests are not blocked.
3. Sends a rate-limited request and captures the Retry-After value.
4. Sleeps 60 seconds, sends another rate-limited request, and asserts
   that the new Retry-After value is strictly less than the first one.
'''

def run_test():
    for _ in range(3):
        response = php_server_get("/")
        assert_response_code_is(response, 200)
        assert_response_body_contains(response, "Request successful")

    time.sleep(3)

    for _ in range(3):
        response = php_server_get("/")

    response = php_server_get("/")
    assert_response_code_is(response, 429)
    assert_response_header_contains(response, "Content-Type", "text")
    assert_response_body_contains(response, "Rate limit exceeded")
    assert "Retry-After" in response.headers, f"Retry-After header missing: {response.headers}"
    first_retry_after = int(response.headers["Retry-After"])
    assert first_retry_after > 0, f"Retry-After should be > 0, got {first_retry_after}"
    assert first_retry_after <= 300, f"Retry-After should be <= 300 (5 min window), got {first_retry_after}"
    print(f"First Retry-After: {first_retry_after}")

    time.sleep(5)

    response = php_server_get("/")
    assert_response_code_is(response, 429)
    assert "Retry-After" in response.headers, f"Retry-After header missing: {response.headers}"
    second_retry_after = int(response.headers["Retry-After"])
    assert second_retry_after > 0, f"Retry-After should be > 0, got {second_retry_after}"
    assert second_retry_after < first_retry_after, f"Retry-After should decrease over time: {second_retry_after} < {first_retry_after}"
    print(f"Second Retry-After: {second_retry_after} (decreased from {first_retry_after})")

if __name__ == "__main__":
    load_test_args()
    run_test()
