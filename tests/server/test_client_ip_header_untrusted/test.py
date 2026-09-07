from testlib import *


def run_test():
    remote = php_server_get("/test")
    assert_response_code_is(remote, 403)
    assert_response_body_contains(remote, "not allowed by config to access this endpoint")

    for headers in [
        {"X-Real-IP": "1.1.1.1"},
        {"X-Forwarded-For": "1.1.1.1"},
        {"X-Real-IP": "1.1.1.1", "X-Forwarded-For": "1.1.1.1"},
    ]:
        response = php_server_get("/test", headers=headers)
        assert_response_code_is(response, 403)
        assert response.text == remote.text, response.text


if __name__ == "__main__":
    load_test_args()
    run_test()
