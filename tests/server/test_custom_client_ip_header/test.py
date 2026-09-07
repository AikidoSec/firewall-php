from testlib import *


def assert_blocked(ip):
    response = php_server_get("/test", headers={"X-Real-IP": ip})
    assert_response_code_is(response, 403)
    assert_response_header_contains(response, "Content-Type", "text")
    assert_response_body_contains(response, f"Your ip ({ip}) is blocked due to: geo restrictions!")


def run_test():
    response = php_server_get("/test", headers={"X-Forwarded-For": "185.141.119.107"})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Something")

    response = php_server_get(
        "/test",
        headers={"X-Real-IP": "invalid", "X-Forwarded-For": "185.141.119.107"},
    )
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Something")

    assert_blocked("185.141.119.107")
    assert_blocked("2a02:6ea0:d213:2205::12")


if __name__ == "__main__":
    load_test_args()
    run_test()
