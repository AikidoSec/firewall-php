from testlib import *

'''
1. Sends a request to a page that prints the PHP version of the running server.
2. Checks that the 'started' event reports that version, not the version of the PHP headers the extension was built against.
'''

def run_test():
    response = php_server_get("/")
    assert_response_code_is(response, 200)

    events = mock_server_get_events()
    assert_event_contains_subset("__root", events[0], {"type": "started", "agent": {"platform": {"version": response.text}}})

if __name__ == "__main__":
    load_test_args()
    run_test()
