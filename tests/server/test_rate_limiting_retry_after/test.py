from testlib import *


def assert_allowed(response):
    assert_response_code_is(response, 200)
    assert "Retry-After" not in response.headers
    assert response.json()["retry_after"] is None


def run_test():
    # Observe the original 429; urllib3 would otherwise wait and retry it.
    s.mount('http://', HTTPAdapter(max_retries=Retry(
        connect=10, backoff_factor=1, respect_retry_after_header=False)))

    # Exercise every identity type, with and without an exact route match.
    for route, delay in [("/", 60), ("/long-window", 600)]:
        for trigger, query in [("ip", ""), ("user", "?user=alice"),
                               ("group", "?user=alice&group=team")]:
            for _ in range(3):
                assert_allowed(php_server_get(route + query))
            for _ in range(2):
                response = php_server_get(route + query)
                assert_response_code_is(response, 429)
                retry_after = int(response.headers["Retry-After"])
                assert 1 <= retry_after <= delay
                decision = response.json()
                assert decision["type"] == "ratelimited"
                assert decision["trigger"] == trigger
                assert decision["retry_after"] == retry_after

        # An unrelated identity must not inherit another request's delay.
        assert_allowed(php_server_get(route + "?user=bob"))

    response = php_server_get("/long-window?user=alice")
    assert_response_code_is(response, 429)
    initial_delay = int(response.headers["Retry-After"])
    time.sleep(2)
    response = php_server_get("/long-window?user=alice")
    assert_response_code_is(response, 429)
    assert 0 < int(response.headers["Retry-After"]) < initial_delay

    response = php_server_get("/?user=blocked-user")
    assert_response_code_is(response, 403)
    assert "Retry-After" not in response.headers
    assert response.json()["retry_after"] is None


if __name__ == "__main__":
    load_test_args()
    run_test()
