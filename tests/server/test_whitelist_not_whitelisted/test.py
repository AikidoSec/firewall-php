import requests
import time
import sys
from testlib import *

'''
Test that should_whitelist_request returns whitelisted=false when none of the
whitelist conditions are met.

No endpoint-level allowlist is configured, the IP is not bypassed, and the IP is
not in any global allowed IP list. All three checks in OnGetWhitelistedStatus
return false, so the default whitelisted=false is returned.
'''


def run_test():
    response = php_server_get("/test", headers={"X-Forwarded-For": "185.245.255.213"})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "whitelisted=false;")
    assert_response_body_contains(response, "type=;")
    assert_response_body_contains(response, "trigger=;")
    assert_response_body_contains(response, "description=;")
    assert_response_body_contains(response, "ip=;")
    assert_response_body_contains(response, "Something!")


if __name__ == "__main__":
    load_test_args()
    run_test()
