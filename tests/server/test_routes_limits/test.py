import requests
import time
import sys
from testlib import *

'''
1. Sets up a simple config.
2. Sends multiple requests to different routes.
3. Waits for the heartbeat event and validates the reporting.
'''

def generate_routes():
    generated_routes = []
    for i in range(6000):
        route = f"/route{i}"
        generated_routes.append(route)
    return generated_routes

def run_test():
    routes = generate_routes()
    for route in routes:
        response = php_server_get(route)
        assert_response_code_is(response, 200)
    
    mock_server_wait_for_new_events(70)
    
    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_started_event_is_valid(events[0])

    
    all_routes = aggregate_field_from_heartbeats("routes", unique_key="path")
    paths = [p["path"] for p in all_routes]
    assert len(paths) == 5000, f"Expected 5000 routes, got {len(paths)}"
    assert routes[0] not in paths, f"Route {routes[0]} should not be in reported paths"
    assert routes[-1] in paths, f"Route {routes[-1]} should be in reported paths"
    
if __name__ == "__main__":
    load_test_args()
    run_test()
