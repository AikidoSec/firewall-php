import requests
import time
import sys
from testlib import *

'''
Test that should_whitelist_request returns whitelisted=true with type=allowlist
when the request IP is found in the global allowed IP list.

The IP 185.245.255.211 is in lists_allowedIPAddresses with description "Manually allowed IPs".
It is not bypassed and no endpoint-level allowlist is configured.
The allowlist check is the third condition in OnGetWhitelistedStatus.
'''


def run_test():
    response = php_server_get("/test", headers={"X-Forwarded-For": "185.245.255.211"})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "whitelisted=true;")
    assert_response_body_contains(response, "type=allowlist;")
    assert_response_body_contains(response, "trigger=ip;")
    assert_response_body_contains(response, "description=IP is part of allowlist: Manually allowed IPs;")
    assert_response_body_contains(response, "ip=185.245.255.211;")
    assert_response_body_contains(response, "Something!")


if __name__ == "__main__":
    load_test_args()
    run_test()
