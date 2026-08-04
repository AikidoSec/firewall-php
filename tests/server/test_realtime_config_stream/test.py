import requests
import time
import sys
from concurrent.futures import ThreadPoolExecutor
from testlib import *

'''
Runs with AIKIDO_FEATURE_SSE enabled, so the agent listens for config updates on a stream
instead of only polling for them.

1. Checks that the route is open with the start config.
2. Checks that the agent connected to the config stream, which proves the feature flag reached
   the agent (the agent is spawned without an environment, so the flag travels over gRPC).
3. Changes the config and checks that the agent fetched it right away, instead of waiting for
   the config polling interval of 1 minute.
4. Checks that the new config reaches the PHP requests. Server APIs serve requests from several
   worker processes, and each one pulls the config from the agent on its own 1 minute interval,
   so the requests are sent in parallel to reach a worker that already pulled it.
'''

# The agent polls for config changes every minute, so a fetch this soon after a config change
# can only have been triggered by a config stream event
STREAM_FETCH_MAX_WAIT = 20

# Each worker process pulls the config from the agent on its own 1 minute interval
CONFIG_PROPAGATION_MAX_WAIT = 120

# Requests sent at once, enough to reach several worker processes of the server under test
PARALLEL_REQUESTS = 6

# A request that takes this long is a failure to report, not something to keep waiting for
REQUEST_TIMEOUT = 30

last_round = "no requests sent yet"


def get_blocked_response():
    """
    Sends the requests in parallel, so they are spread over the worker processes instead of all
    being served by whichever single worker happens to be idle. Returns the response of a worker
    that blocked the route, or None while no worker has the new config yet.
    """
    global last_round
    url = f"http://localhost:{get_php_port()}/somethingVerySpecific"

    def get():
        try:
            return requests.get(url, timeout=REQUEST_TIMEOUT)
        except requests.RequestException as e:
            return e

    with ThreadPoolExecutor(max_workers=PARALLEL_REQUESTS) as pool:
        results = list(pool.map(lambda _: get(), range(PARALLEL_REQUESTS)))

    last_round = [r if isinstance(r, Exception) else r.status_code for r in results]
    blocked = [r for r in results if not isinstance(r, Exception) and r.status_code == 403]
    return blocked[0] if blocked else None


def wait_for_blocked_response():
    deadline = time.time() + CONFIG_PROPAGATION_MAX_WAIT
    while True:
        response = get_blocked_response()
        if response is not None:
            return response
        if time.time() >= deadline:
            return None
        time.sleep(2)


def mock_server_state():
    return (f"stream connections: {mock_server_get_stream_connections()}, "
            f"config fetches: {mock_server_get_config_fetch_count()}, "
            f"last responses: {last_round}")


def run_test():
    response = php_server_get("/somethingVerySpecific")
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Something")

    connected = wait_until(lambda: mock_server_get_stream_connections() > 0, 30)
    assert connected is not None, f"Agent did not connect to the config stream ({mock_server_state()})"

    config_fetches = mock_server_get_config_fetch_count()
    mock_server_set_config_file("change_config_allowed_ip.json")

    fetched_in = wait_until(lambda: mock_server_get_config_fetch_count() > config_fetches, STREAM_FETCH_MAX_WAIT)
    assert fetched_in is not None, (f"Agent did not fetch the new config within {STREAM_FETCH_MAX_WAIT}s "
                                    f"of it changing ({mock_server_state()})")
    print(f"Agent fetched the new config {fetched_in:.1f}s after it changed")

    started_waiting = time.time()
    response = wait_for_blocked_response()
    assert response is not None, (f"New config did not reach the PHP requests within "
                                  f"{CONFIG_PROPAGATION_MAX_WAIT}s ({mock_server_state()})")
    print(f"New config reached the PHP requests {time.time() - started_waiting:.1f}s after it changed")

    assert_response_code_is(response, 403)
    assert_response_header_contains(response, "Content-Type", "text")
    assert_response_body_contains(response, "is blocked due to: not allowed by config to access this endpoint!")


if __name__ == "__main__":
    load_test_args()
    run_test()
