import requests
import time
import sys
from testlib import *

'''
Checks that simple strings with a trailing comma (e.g. "option1,")
are not flagged as SQL injection.
'''

def check_not_blocked(title):
    response = php_server_post("/testDetection", {"title": title})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Query executed!")

def run_test():
    # "p," occurs in the query as a substring of "zip, country"
    check_not_blocked("p,")

if __name__ == "__main__":
    load_test_args()
    run_test()
