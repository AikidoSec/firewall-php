import requests
import time
import sys
from testlib import *

'''
Test that should_block_request returns a valid object when IP is bypassed.

The IP 185.245.255.211 is in the global allowedIPAddresses (bypassed).
The endpoint only allows 185.245.255.212 via endpoint-level allowedIPAddresses.

A request from the bypassed IP should not crash when accessing
properties on the return value of should_block_request().
'''


def run_test():
    response = php_server_get("/somethingVerySpecific", headers={"X-Forwarded-For": "185.245.255.211"})
    assert_response_code_is(response, 200)
    assert_response_body_not_contains(response, "Decision is null!")
    assert_response_body_contains(response, "Something!")


if __name__ == "__main__":
    load_test_args()
    run_test()
