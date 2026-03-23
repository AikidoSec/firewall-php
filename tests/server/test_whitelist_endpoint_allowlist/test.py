import requests
import time
import sys
from testlib import *

'''
Test that should_whitelist_request returns whitelisted=true with type=endpoint-allowlist
when the endpoint has an IP allowlist configured and the request IP is not in it.

The endpoint /test only allows 185.245.255.212 via endpoint-level allowedIPAddresses.
The IP 185.245.255.211 is globally bypassed so auto_block_request does not exit the script.
In OnGetWhitelistedStatus, the endpoint-allowlist check comes before the bypass check,
so whitelisted=true with type=endpoint-allowlist is returned.
'''


def run_test():
    response = php_server_get("/test", headers={"X-Forwarded-For": "185.245.255.212"})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "whitelisted=true;")
    assert_response_body_contains(response, "type=endpoint-allowlist;")
    assert_response_body_contains(response, "trigger=ip;")
    assert_response_body_contains(response, "description=IP is configured in the route&#39;s allowlist;")
    assert_response_body_contains(response, "ip=185.245.255.212;")
    assert_response_body_contains(response, "Something!")


if __name__ == "__main__":
    load_test_args()
    run_test()
