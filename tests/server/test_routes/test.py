import requests
import time
import sys
from testlib import *

'''
1. Sets up a simple config.
2. Sends multiple requests to different routes.
3. Waits for the heartbeat event and validates the reporting.
'''

routes = {
    "/",
    "/test",
    "/api/info/12345",
    "/api/info/12345678",
    "/api/secret/CnJ4DunhYfv2db6T1FRfciRBHtlNKOYrjoz",
    "/api/secret/hIofuWBifkJI5iVsSNKKKDpBfmMqJJwuXMxau6AS8WZaHVLDAMeJXo3BwsFyrIIm",
    "/api/uuid/2aa11254-7c5a-4427-b819-a85063b157ca",
    "/api/uuid/52f80bbb-a3aa-44bc-b980-e51db0785fe3",
    "/api/uuid/6d380fc0-143e-4c48-b456-82ac4592fe06",
    "/posts/2023-05-01",
    "/posts/2024-05-01",
    "/posts/2022-03-28",
    "/posts/1994-09-01",
    "/block/2001:2:ffff:ffff:ffff:ffff:ffff:ffff",
    "/files/098f6bcd4621d373cade4e832627b4f6",
    "/files/a94a8fe5ccb19ba61c4c0873d391e987982fbbd3",
    "/posts/ECBCDD2C-A441-4846-B5AC-0083D347FDF2"
}

def run_test():
    for route in routes:
        for _ in range(10):
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
