import requests
import time
import sys
from testlib import *

'''
Test that should_whitelist_request returns whitelisted=true with type=allowlist
when the request IP is found in the global allowed IP list.

Both 185.245.255.211 and 185.245.255.214 are in lists_allowedIPAddresses with
description "Manually allowed IPs". Multiple requests from both IPs should all
return whitelisted=true. No bypass and no endpoint-level allowlist is configured.
'''


def assert_whitelisted_for_ip(ip):
    response = php_server_get("/test", headers={"X-Forwarded-For": ip})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "whitelisted=true;")
    assert_response_body_contains(response, "type=allowlist;")
    assert_response_body_contains(response, "trigger=ip;")
    assert_response_body_contains(response, "description=IP is part of allowlist: Manually allowed IPs;")
    assert_response_body_contains(response, f"ip={ip};")
    assert_response_body_contains(response, "Something!")


def run_test():
    assert_whitelisted_for_ip("185.245.255.211")
    assert_whitelisted_for_ip("185.245.255.214")
    assert_whitelisted_for_ip("185.245.255.211")
    assert_whitelisted_for_ip("185.245.255.214")
    assert_whitelisted_for_ip("185.245.255.214")
    assert_whitelisted_for_ip("185.245.255.211")


if __name__ == "__main__":
    load_test_args()
    run_test()
