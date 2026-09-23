from testlib import *


def custom_events():
    return [event for event in mock_server_get_events() if event.get("type") == "custom"]


def run_test():
    client_ip = "1.2.3.4"
    headers = {
        "User-Agent": "custom-event-test",
        "X-Forwarded-For": client_ip,
    }

    response = php_server_get("/?token=secret&email=user@example.com", headers=headers)
    assert_response_code_is(response, 403)
    assert_response_body_contains(response, "User blocked after events were tracked")

    assert wait_until(lambda: len(custom_events()) == 2, 10) is not None

    events_by_name = {event["name"]: event for event in custom_events()}
    login_failed = events_by_name["user.login_failed"]
    assert_event_contains_subset(
        "__root",
        login_failed,
        {
            "request": {
                "method": "GET",
                "ipAddress": client_ip,
                "userAgent": "custom-event-test",
                "source": "php",
                "route": "/",
            },
            "agent": {"library": "firewall-php"},
        },
    )
    assert isinstance(login_failed["time"], int)
    assert "url" not in login_failed["request"]
    assert "user" not in login_failed

    login_succeeded = events_by_name["user.login_succeeded"]
    assert login_succeeded["user"] == {"id": "user-1", "name": "Jane Doe"}


if __name__ == "__main__":
    load_test_args()
    run_test()
