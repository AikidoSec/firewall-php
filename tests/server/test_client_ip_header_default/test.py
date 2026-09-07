from testlib import *


def run_test():
    response = php_server_get("/test", headers={
        "X-Forwarded-For": "8.8.8.8",
        "X-Real-IP": "1.1.1.1",
    })
    assert_response_code_is(response, 403)
    assert_response_body_contains(response, "Your ip (8.8.8.8)")

    response = php_server_get("/test", headers={"X-Forwarded-For": "1.1.1.1"})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Request successful!")


if __name__ == "__main__":
    load_test_args()
    run_test()
