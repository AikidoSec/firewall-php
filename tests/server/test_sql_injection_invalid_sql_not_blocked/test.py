import time
from testlib import *

"""
With AIKIDO_BLOCK=1 but AIKIDO_BLOCK_INVALID_SQL=0, queries that only fail tokenization
(result 3) are not blocked and no attack event is reported.
"""


def run_test():
    response = php_server_get("/testDetection?id=1+%2F*")
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Error: SQLSTATE[HY000]")

if __name__ == "__main__":
    load_test_args()
    run_test()
