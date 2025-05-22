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
    generated_routes = set()
    while len(generated_routes) < 6000:
        route_len = random.randint(1, 20)  # Random route length between 1-20 segments
        route_parts = []
        for j in range(route_len):
            part_len = random.randint(3, 15)  # Random segment length between 3-15 chars
            part = ''.join(random.choices('abcdefghijklmnopqrstuvwxyz0123456789-_', k=part_len))
            route_parts.append(part)
        route = '/' + '/'.join(route_parts)
        generated_routes.add(route)
    return list(generated_routes)

def run_test():
    routes = generate_routes()
    for route in routes:
        response = php_server_get(route)
        assert_response_code_is(response, 200)
    
    mock_server_wait_for_new_events(70)
    
    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_started_event_is_valid(events[0])

    paths = [p["path"] for p in events[1]["routes"]]
    assert len(paths) == 5000, f"Expected 5000 routes, got {len(paths)}"
    assert routes[0] not in paths, f"Route {routes[0]} should not be in reported paths"
    assert routes[-1] in paths, f"Route {routes[1]} should be in reported paths"
    
if __name__ == "__main__":
    load_test_args()
    run_test()
