from testlib import *

"""
Invalid SQL (failed tokenization) with user input is blocked when AIKIDO_BLOCK=1 and
AIKIDO_BLOCK_INVALID_SQL=1. Uses PDO SQLite and an unclosed block comment in the query.
"""


def run_test():
    response = php_server_get("/testDetection?id=1+%2F*")
    assert_response_code_is(response, 500)
    assert_response_body_contains(response, "")

    mock_server_wait_for_new_events(5)

    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_started_event_is_valid(events[0])
    assert_event_contains_subset_file(events[1], "expect_detection_blocked.json")


if __name__ == "__main__":
    load_test_args()
    run_test()
