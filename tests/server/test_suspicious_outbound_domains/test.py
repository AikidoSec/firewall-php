from testlib import (
    apply_config,
    assert_event_contains_subset,
    get_mock_port,
    load_test_args,
    mock_server_get_events,
    php_server_get,
    wait_until,
)


def detections():
    return [event for event in mock_server_get_events() if event["type"] == "detected_attack"]


def run_test():
    port = get_mock_port()
    response = php_server_get(f"/blocked?port={port}")
    assert response.status_code == 500
    assert "blocked an outbound connection" in response.text

    expected = {
        "callback": ("probe.oast.live", "oast.live"),
        "ftp": ("probe.oast.live", "oast.live"),
        "rebinding": ("probe.rbndr.us", "rbndr.us"),
        "wildcard": ("test.sslip.io", "sslip.io"),
    }
    response = php_server_get(f"/lookalike?port={port}")
    assert response.status_code == 200
    assert "config_fetch_count" in response.json()

    for route in expected:
        response = php_server_get(f"/{route}?port={port}")
        assert response.status_code == 500
        assert "blocked a server-side request forgery" in response.text

    assert wait_until(lambda: len(detections()) >= len(expected), 10) is not None
    events = detections()
    assert len(events) == len(expected)
    by_route = {event["request"]["route"]: event for event in events}
    for route, (hostname, domain) in expected.items():
        assert_event_contains_subset("__root", by_route[f"/{route}"], {
            "request": {"method": "GET", "route": f"/{route}"},
            "attack": {
                "kind": "ssrf",
                "blocked": True,
                "operation": "curl_exec",
                "source": "outbound URL",
                "payload": hostname,
                "stack": "index.php",
                "metadata": {
                    "reason": "suspicious_outbound_domain",
                    "matchedDomain": domain,
                    "hostname": hostname,
                    "port": str(port),
                },
            },
        })

    apply_config("disable_blocking.json")
    response = php_server_get(f"/rebinding?port={port}")
    assert response.status_code == 200
    assert "config_fetch_count" in response.json()

    response = php_server_get(f"/lookalike?port={port}")
    assert response.status_code == 200
    assert "config_fetch_count" in response.json()
    assert wait_until(lambda: len(detections()) > len(expected), 10) is not None
    events = detections()
    assert len(events) == len(expected) + 1
    assert_event_contains_subset("__root", events[-1], {
        "request": {"route": "/lookalike"},
        "attack": {"kind": "ssrf", "blocked": False, "payload": "notoast.live"},
    })


if __name__ == "__main__":
    load_test_args()
    run_test()
