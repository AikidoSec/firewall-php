import requests
import time
import sys
from testlib import *

'''
Checks that simple strings with a trailing comma (e.g. "option1,")
are not flagged as SQL injection.
'''

def check_not_blocked(input):
    response = php_server_post("/testDetection", {"name": "John", "input": input})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Query executed!")

def run_test():
    check_not_blocked("option1,")
    check_not_blocked("hello,")
    check_not_blocked("test,")
    check_not_blocked("value1, value2,")
    check_not_blocked("foo, bar,")

if __name__ == "__main__":
    load_test_args()
    run_test()
