import requests
import time
import sys
from testlib import *

'''
Checks that common SQL strings like "is not", "not in" etc. do not
trigger false positive SQL injection detections.
'''

def check_not_blocked(input):
    response = php_server_post("/testDetection", {"name": "John", "input": input})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Query executed!")

def run_test():
    check_not_blocked("is not")
    check_not_blocked("not in")
    check_not_blocked("TIME ZONE")
    check_not_blocked("IS NOT")
    check_not_blocked(":n")

if __name__ == "__main__":
    load_test_args()
    run_test()
