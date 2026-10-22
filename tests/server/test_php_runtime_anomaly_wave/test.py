import requests
import time
import sys
from testlib import *

'''
Requests that hit PHP fatal, out-of-memory or timeout errors, or that carry the session upload progress field,
are signals of attempts to exploit CVE-2026-109618 / CVE-2026-109619.
1. For each signal, send 10 such requests from a new IP and check that one attack wave event is reported,
   with the "php_runtime_anomalies" reason and the signal as the reason of the samples.
2. Send 10 more requests from an IP that was already reported, no new event should be sent.
3. Send 10 requests from a bypassed IP, no event should be sent.
Out-of-memory errors are only checked on PHP versions that are still vulnerable: the patched versions terminate
the process on out-of-memory errors, before request shutdown runs. The PHP built-in server exits on those errors,
so that signal is skipped there.
'''

THRESHOLD = 10


def php_survives_out_of_memory():
    if mock_server_get_platform_name() == "cli-server":
        return False
    try:
        response = requests.get(f"http://localhost:{get_php_port()}/?oom=1", headers={"X-Forwarded-For": "5.8.19.49"}, timeout=30)
    except requests.exceptions.ConnectionError:
        return False
    # a reverse proxy answers with a gateway error when the PHP worker was terminated
    return response.status_code not in (502, 503, 504)


def post_upload_progress(ip):
    return requests.post(
        f"http://localhost:{get_php_port()}/upload",
        data={"PHP_SESSION_UPLOAD_PROGRESS": "exploit"},
        files={"file": ("test.txt", b"test")},
        headers={"X-Forwarded-For": ip},
    )


def send_requests(signal, ip):
    for _ in range(THRESHOLD):
        if signal == "session_upload_progress":
            post_upload_progress(ip)
        else:
            query = {"memory_exhausted": "oom", "execution_timeout": "timeout", "fatal_error": "exception"}[signal]
            php_server_get(f"/?{query}=1", headers={"X-Forwarded-For": ip})


def assert_runtime_anomaly_wave(event, ip, signal):
    assert_event_contains_subset("__root", event, {
        "type": "detected_attack_wave",
        "request": {"ipAddress": ip},
        "attack": {"metadata": {"reason": "php_runtime_anomalies"}},
        "agent": {"library": "firewall-php"},
    })
    samples = json.loads(event["attack"]["metadata"]["samples"])
    assert len(samples) > 0, "Samples not found in metadata"
    for sample in samples:
        assert sample["reason"] == signal, f"Expected reason {signal} in sample {sample}"


def run_test():
    signals = [
        ("memory_exhausted", "5.8.19.40"),
        ("execution_timeout", "5.8.19.41"),
        ("fatal_error", "5.8.19.42"),
        ("session_upload_progress", "5.8.19.43"),
    ]
    # the platform is known only after the agent sent the "started" event
    php_server_get("/")
    wait_until(lambda: mock_server_get_platform_name() != "", 30)
    if not php_survives_out_of_memory():
        print("Skipping out-of-memory checks: PHP terminates the process on out-of-memory errors")
        signals = [s for s in signals if s[0] != "memory_exhausted"]

    expected_events = 1
    for signal, ip in signals:
        send_requests(signal, ip)
        mock_server_wait_for_new_events(10)

        expected_events += 1
        events = mock_server_get_events()
        assert_events_length_is(events, expected_events)
        assert_started_event_is_valid(events[0])
        assert_runtime_anomaly_wave(events[-1], ip, signal)

    # no new event for an IP that was already reported
    send_requests("fatal_error", "5.8.19.42")
    mock_server_wait_for_new_events(10)
    assert_events_length_is(mock_server_get_events(), expected_events)

    # no event for a bypassed IP
    send_requests("fatal_error", "5.8.19.44")
    mock_server_wait_for_new_events(10)
    assert_events_length_is(mock_server_get_events(), expected_events)


if __name__ == "__main__":
    load_test_args()
    run_test()
