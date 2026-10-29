from testlib import *


def run_test():
    response = php_server_get("/test")
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "Something")
    assert mock_server_get_token() == "foobar"


if __name__ == "__main__":
    load_test_args()
    run_test()
