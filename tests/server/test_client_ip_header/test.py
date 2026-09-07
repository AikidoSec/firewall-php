from testlib import *


def run_test():
    # The endpoint allows only 1.1.1.1, so the blocking response exposes the
    # resolved client IP. Capture the real connection's response for fallbacks.
    remote = php_server_get("/test")
    assert_response_code_is(remote, 403)
    assert_response_body_contains(remote, "not allowed by config to access this endpoint")

    cases = [
        ("8.8.8.8", "8.8.8.8"),
        ("2001:4860:4860::8888", "2001:4860:4860::8888"),
        ("8.8.8.8:443", "8.8.8.8"),
        ("[2001:4860:4860::8888]:443", "2001:4860:4860::8888"),
        ("invalid, 10.0.0.1, 8.8.8.8, 9.9.9.9", "8.8.8.8"),
    ]
    for header, expected_ip in cases:
        response = php_server_get("/test", headers={
            "X-Real-IP": header,
            "X-Forwarded-For": "1.1.1.1",
        })
        assert_response_code_is(response, 403)
        assert_response_body_contains(response, f"Your ip ({expected_ip})")

    # Alternate success and fallback requests to catch cached IPs in workers.
    for header in [None, "", "invalid", "10.0.0.1, 127.0.0.1", "::1"]:
        allowed = php_server_get("/test", headers={
            "x-real-ip": "1.1.1.1",
            "X-Forwarded-For": "8.8.8.8",
        })
        assert_response_code_is(allowed, 200)
        assert_response_body_contains(allowed, "Request successful!")

        headers = {"X-Forwarded-For": "1.1.1.1"}
        if header is not None:
            headers["X-Real-IP"] = header
        response = php_server_get("/test", headers=headers)
        assert_response_code_is(response, 403)
        assert response.text == remote.text, response.text


if __name__ == "__main__":
    load_test_args()
    run_test()
