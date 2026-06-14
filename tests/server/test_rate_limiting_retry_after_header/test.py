import requests
import time
import sys
from testlib import *

'''
Window: 2 minutes (2 buckets of 1 minute each), max 3 requests.

Phase 1: Basic Retry-After + countdown
  - Send 3 requests (OK), trigger rate limiting, check Retry-After <= 120.
  - Sleep 5s, trigger again, assert Retry-After decreased.

Phase 2: Retry-After stays accurate across bucket evictions
  - Wait for window to expire (~120s). Rate limit resets.
  - Send 1 request at T=0 (bucket 0 = 1, total = 1).
  - Sleep 65s so bucket advances. Send 1 request (bucket 1 = 1, total = 2).
  - Sleep 65s. Bucket 0 is evicted, total drops to 1, window survives.
  - Send 2 more requests to push total back to 3 → rate limited again.
  - Assert Retry-After > 10 and <= 120.

Phase 3: Full reset
  - Wait for window to expire. Trigger rate limiting again.
  - Assert Retry-After resets back near 120.
'''

def run_test():
    # --- Phase 1: basic countdown ---
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
    assert first_retry_after <= 120, f"Retry-After should be <= 120 (2 min window), got {first_retry_after}"
    print(f"Phase 1a - First Retry-After: {first_retry_after}")

    time.sleep(5)

    response = php_server_get("/")
    assert_response_code_is(response, 429)
    assert "Retry-After" in response.headers, f"Retry-After header missing: {response.headers}"
    second_retry_after = int(response.headers["Retry-After"])
    assert second_retry_after > 0, f"Retry-After should be > 0, got {second_retry_after}"
    assert second_retry_after < first_retry_after, f"Retry-After should decrease: {second_retry_after} < {first_retry_after}"
    print(f"Phase 1b - Second Retry-After: {second_retry_after} (decreased from {first_retry_after})")

    # --- Phase 2: multi-bucket eviction (catches CreatedAt bug) ---
    # Wait for window to fully expire
    time.sleep(120)

    # T=0: send 1 request into bucket 0
    response = php_server_get("/")
    assert_response_code_is(response, 200)

    # Wait for bucket advance (~65s)
    time.sleep(65)

    # T=65: send 1 request into bucket 1 (total=2, spread across 2 buckets)
    response = php_server_get("/")
    assert_response_code_is(response, 200)

    # Wait for next tick to evict bucket 0 (~65s)
    # After eviction: bucket 0 (count=1) is dropped, total=1, window survives via bucket 1
    time.sleep(65)

    # T=130: send 2 requests to push total back to 3 (re-triggering rate limit)
    response = php_server_get("/")
    assert_response_code_is(response, 200)
    response = php_server_get("/")
    assert_response_code_is(response, 200)

    # Now total=3, next request should be rate limited
    response = php_server_get("/")
    assert_response_code_is(response, 429)
    assert "Retry-After" in response.headers, f"Retry-After header missing: {response.headers}"
    eviction_retry_after = int(response.headers["Retry-After"])
    print(f"Phase 2 - Retry-After after bucket eviction: {eviction_retry_after} (would be 1 without CreatedAt fix)")
    assert eviction_retry_after > 10, f"Retry-After should be > 10 after eviction (would be 1 without fix), got {eviction_retry_after}"
    assert eviction_retry_after <= 120, f"Retry-After should be <= 120 (2 min window), got {eviction_retry_after}"

    # --- Phase 3: full reset ---
    time.sleep(120)

    for _ in range(3):
        response = php_server_get("/")
        assert_response_code_is(response, 200)

    time.sleep(3)

    for _ in range(3):
        response = php_server_get("/")

    response = php_server_get("/")
    assert_response_code_is(response, 429)
    assert "Retry-After" in response.headers, f"Retry-After header missing: {response.headers}"
    reset_retry_after = int(response.headers["Retry-After"])
    assert reset_retry_after > 100, f"Retry-After should reset near 120 after full window expiry, got {reset_retry_after}"
    print(f"Phase 3 - Retry-After after full reset: {reset_retry_after}")

if __name__ == "__main__":
    load_test_args()
    run_test()
