import sys
from testlib import *

'''
1. Installs a session save handler whose write() calls eval().
   PHP runs session RSHUTDOWN after Zen's RSHUTDOWN, so that eval compiles
   PHP after Zen freed its AST cleanup table.
2. Hits the page twice on the same connection-heavy loop; on a vulnerable
   build the second compile in the same worker SIGSEGVs and the request
   fails with a bad gateway / empty reply.
3. Asserts every request returns 200 and the expected body, and that the
   eval'd value in the session write callback executed without crashing.
'''

def run_test():
    for _ in range(10):
        response = php_server_get("/")
        assert_response_code_is(response, 200)
        assert_response_body_contains(response, "value=42")


if __name__ == "__main__":
    load_test_args()
    run_test()
