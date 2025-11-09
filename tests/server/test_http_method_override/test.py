import requests
import time
import sys
from testlib import *
import shutil
import os
import json

'''
Tests HTTP method override functionality when Symfony HTTP Foundation is present.

1. Verifies _method query parameter overrides (POST → PUT, DELETE, GET)
2. Verifies X-HTTP-Method-Override header overrides (POST → PATCH, PUT, GET)
3. Tests case-insensitive header handling (x-http-Method-Override)
4. Validates that rate limiting applies to the overridden method, not the original POST
   - Configures 5 GET requests/minute limit
   - Sends 5 GET requests (allowed)
   - Sends POST requests with method override to GET (rate limited as GET)
'''

COMPOSER_LOCK_FULL_PATH = os.path.join(os.path.dirname(__file__), "../composer.lock")

def run_test():
  
    response = php_server_post("/?_method=PUT", {})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Method: PUT")
    
    response = php_server_post("/?_method=DELETE", {})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Method: DELETE")

    response = php_server_post("/", {}, headers={"X-HTTP-Method-Override": "PATCH"})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Method: PATCH")

    response = php_server_post("/", {}, headers={"X-HTTP-Method-Override": "PUT"})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Method: PUT")

    response = php_server_post("/", {})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Method: POST")

    response = php_server_post("/?_method=put", {})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Method: PUT")

    for _ in range(5):
        response = php_server_get("/")
        assert_response_code_is(response, 200)
        assert_response_body_contains(response, "Method: GET")

    time.sleep(10)
    response = php_server_post("/", {}, headers={"X-HTTP-Method-Override": "GET"})
    assert_response_code_is(response, 429)

    response = php_server_post("/", {}, headers={"x-http-Method-Override": "GET"})
    assert_response_code_is(response, 429)


    response = php_server_post("/?_method=GET", {})
    assert_response_code_is(response, 429)

    mock_server_wait_for_new_events(70)

    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_started_event_is_valid(events[0])
    assert_event_contains_subset_file(events[1], "expect_method_overrides.json")

def create_composer_lock():
    package_data = {"packages": [{"name": "symfony/http-foundation", "version": "v7.3.6"}]}
    with open(COMPOSER_LOCK_FULL_PATH, "w") as f:
        json.dump(package_data, f)

def delete_composer_lock():
    os.remove(COMPOSER_LOCK_FULL_PATH)

if __name__ == "__main__":
    load_test_args()

    try:
        create_composer_lock()
        run_test()
    finally:
        delete_composer_lock()
