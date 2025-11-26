import time
from testlib import *

'''
E2E test for router param matching functionality.

1. PHP code (index.php) registers custom param matchers on each request.
2. This test sends requests that should match both custom and default matchers.
3. Waits for the heartbeat event and validates routes reporting.
'''

routes = [
    "/test",
    "/posts/2023-05-01",                         # default :date matcher
    "/posts/aikido-123",                         # custom :tenant matcher
    "/blog/aikido-123/aikido-foo-123-bar",       # custom :tenant + :slug matchers
]


def run_test():
    # Hit each route 5 times to match expect_routes.json
    for route in routes:
        for _ in range(5):
            response = php_server_get(route)
            assert_response_code_is(response, 200)

    mock_server_wait_for_new_events(70)

    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_started_event_is_valid(events[0])
    assert_event_contains_subset_file(events[1], "expect_routes.json")


if __name__ == "__main__":
    load_test_args()
    run_test()


