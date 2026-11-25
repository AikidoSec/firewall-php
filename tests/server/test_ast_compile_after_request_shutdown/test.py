import sys
from testlib import *

'''
Requires a script from a header callback, which compiles PHP after Zen's
RSHUTDOWN, and asserts the worker keeps serving requests.
'''

def run_test():
    for _ in range(10):
        response = php_server_get("/")
        assert_response_code_is(response, 204)


if __name__ == "__main__":
    load_test_args()
    run_test()
