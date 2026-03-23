import requests
import time
import sys
from testlib import *

'''
Test that should_whitelist_request returns whitelisted=true with type=bypassed
when the request IP is in the global bypass list (allowedIPAddresses).

The IP 185.245.255.211 is globally bypassed via top-level allowedIPAddresses.
No endpoint-level allowlist is configured, so the endpoint-allowlist check does not
trigger. The bypass check is the second condition in OnGetWhitelistedStatus.
'''


def run_test():
    response = php_server_get("/test", headers={"X-Forwarded-For": "185.245.255.211"})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "whitelisted=true;")
    assert_response_body_contains(response, "type=bypassed;")
    assert_response_body_contains(response, "trigger=ip;")
    assert_response_body_contains(response, "description=IP is configured in the firewall bypass list;")
    assert_response_body_contains(response, "ip=185.245.255.211;")
    assert_response_body_contains(response, "Something!")


if __name__ == "__main__":
    load_test_args()
    run_test()
