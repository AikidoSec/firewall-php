import requests
import time
import sys
import json
import os
import random
import string
import datetime
import ipaddress

test_name = ""
php_port = 0
mock_port = 0
benchmarks = []

from requests.adapters import HTTPAdapter, Retry
s = requests.Session()
retries = Retry(connect=10,
                backoff_factor=1)

s.mount('http://', HTTPAdapter(max_retries=retries))

def load_test_args():
    global test_name, php_port, mock_port, initial_mock_port
    php_port = int(sys.argv[1])
    mock_port = int(sys.argv[2])
    test_name = sys.argv[3]
    print(f"Loaded test args: test_name={test_name}, php_port={php_port}, mock_port={mock_port}")

def get_mock_port():
    return mock_port

def get_php_port():
    return php_port

def localhost_get_request(port, route="", headers={}, benchmark=False):
    global benchmarks, s

    start_time = datetime.datetime.now()

    r = s.get(f"http://localhost:{port}{route}", headers=headers)

    end_time = datetime.datetime.now()
    delta = end_time - start_time
    elapsed_ms = delta.total_seconds() * 1000

    if benchmark:
        benchmarks.append(elapsed_ms)

    time.sleep(0.001)
    return r

def localhost_post_request(port, route, data, headers={}, benchmark=False, domain="localhost"):
    global benchmarks, s

    start_time = datetime.datetime.now()

    r = s.post(f"http://{domain}:{port}{route}", json=data, headers=headers)
    end_time = datetime.datetime.now()
    delta = end_time - start_time
    elapsed_ms = delta.total_seconds() * 1000

    if benchmark:
        benchmarks.append(elapsed_ms)

    time.sleep(0.001)
    return r

def php_server_get(route="", headers={}, benchmark=False):
    return localhost_get_request(php_port, route, headers, benchmark)

def php_server_post(route, data, headers={}, benchmark=False):
    return localhost_post_request(php_port, route, data, headers, benchmark)

def php_server_post_custom_domain(domain, route, data, headers={}, benchmark=False):
    return localhost_post_request(php_port, route, data, headers, benchmark, domain)

def mock_server_get(route=""):
    return localhost_get_request(mock_port, route, False)

def mock_server_post(route, data):
    return localhost_post_request(mock_port, route, data, False)

def mock_server_get_events():
    return mock_server_get("/mock/events").json()

def mock_server_set_config(config):
    return mock_server_post("/mock/config", config)

def mock_server_set_config_file(config_file):
    config = None
    with open(config_file, 'r') as f:
        config = json.load(f)
    return mock_server_post("/mock/config", config)

def apply_config(config_file):
    mock_server_set_config_file(config_file)
    time.sleep(120)

def assert_events_length_is(events, length):
    assert isinstance(events, list), "Error: Events is not a list."
    assert len(events) == length, f"Error: Events list contains {len(events)} elements and not {length} elements."

subset_keys_contains_check = ["stack"]
subset_keys_version_check = {
    "stack": "8.1.0",
}

def assert_event_contains_subset(event_subset_key, event, event_subset, dry_mode=False):
    """
    Recursively checks that all keys and values in the subset JSON exist in the event JSON
    and have the same values. If a key in the subset is a list, all its elements must exist in the
    corresponding list in the event.

    :param event: The event JSON dictionary
    :param subset: The subset JSON dictionary
    :raises AssertionError: If the subset is not fully contained within the event
    """
    def result(assertion_error):
        if dry_mode:
            return False
        raise assertion_error

    print(f"Searching {event_subset} in {event} (dry_mode = {dry_mode})...")

    if event is None:
        print(f"Event is None!")
        return False

    if isinstance(event_subset, dict):
        found_all_keys = True
        for key, value in event_subset.items():
            if key not in event:
                return result(AssertionError(f"Key '{key}' not found in '{event}'."))
            if not assert_event_contains_subset(key, event[key], value, dry_mode):
                found_all_keys = False
        return found_all_keys
    elif isinstance(event_subset, list):
        if not isinstance(event, list):
            return result(AssertionError(f"Expected a list in event but found '{event}'."))
        for index, event_subset_item in enumerate(event_subset):
            found_item = False
            for event_item in event:
                if assert_event_contains_subset("__index" + str(index), event_item, event_subset_item, dry_mode=True):
                    found_item = True
                    break
            if not found_item:
                return result(AssertionError(f"Item '{event_subset_item}' not found in {event}."))
    else:
        if event_subset_key in subset_keys_version_check:
            php_version = mock_server_get_php_version()
            if mock_server_get_php_version() < subset_keys_version_check[event_subset_key]:
                print(f"PHP version {php_version} is too old for checking key {event_subset_key}. Skipping check...")
                return True
                
        if event_subset_key in subset_keys_contains_check:
            # Checking if subset field is contained in event field
            if event_subset not in event:
                return result(AssertionError(f"Value mismatch: {event_subset} not in {event}"))
        else:
            # Checking if subset field is equal to event field
            if event_subset != event:
                return result(AssertionError(f"Value mismatch: {event_subset} != {event}"))

    return True

def assert_event_contains_subset_file(event, event_subset_file):
    event_subset = None
    with open(event_subset_file, 'r') as file:
        event_subset = json.load(file)
    assert event_subset
    assert_event_contains_subset("__root",event, event_subset)

def assert_started_event_is_valid(event):
    assert_event_contains_subset("__root", event, {"type": "started", "agent": { "library": "firewall-php" } })

def assert_detection_event_is_valid(event):
    assert_event_contains_subset("__root", event, {"type": "detected_attack", "agent": { "library": "firewall-php" } })

def assert_response_code_is(response, status_code):
    assert response.status_code == status_code, f"Status codes are not the same: {response.status_code} vs {status_code}"

def assert_response_header_contains(response, header, value):
    assert header in response.headers, f"Header '{header}' is not part of response headers: {response.headers}"
    assert value in response.headers[header], f"Header '{header}' does not contain '{value}' but '{response.headers[header]}'"

def assert_response_body_contains(response, text):
    assert text in response.text, f"Test '{text}' is not part of response body: {response.text}"

def assert_is_valid_ip(ip_address):
    try:
        ipaddress.ip_address(ip_address)
        return True
    except ValueError:
        pass
    assert False, "Exception in assert_is_valid_ip for IP address: {ip_address}"

def mock_server_wait_for_new_events(max_wait_time):
    initial_number_of_events = len(mock_server_get_events())
    while max_wait_time > 0:
        if len(mock_server_get_events()) > initial_number_of_events:
            return True
        time.sleep(5)
        max_wait_time -= 5

    return False


def generate_random_string(length):
    return ''.join(random.choices(string.ascii_letters + string.digits, k=length))

def generate_json(size_kb):
    data = {}
    size_bytes = size_kb * 1024  # Convert megabytes to bytes
    current_size = 0

    while current_size < size_bytes:
        key = generate_random_string(random.randint(5, 20))
        value = generate_random_string(random.randint(5, 30))

        data[key] = value

        current_size = len(json.dumps(data))

    return json.dumps(data)

def is_aikido_installed():
    for entry in os.listdir("/opt"):
        if entry.startswith("aikido"):
            return True
    return False

def benchmark_warmup():
    for _ in range(1000):
        php_server_post("/test", {})

def benchmark_store_results():
    global benchmarks
    benchmarks.sort()
    benchmark_suffix = "without_aikido"
    if is_aikido_installed():
        benchmark_suffix = "with_aikido"
    with open(f"{test_name}_{benchmark_suffix}.txt", "w") as f:
        f.write(f"p50 - {benchmarks[int(len(benchmarks)/2)]} ms")

def add_to_hosts_file(hostname, ip):
    hosts_file = '/etc/hosts'
    entry = f"{ip}\t{hostname}\n"

    with open(hosts_file, 'r') as file:
        content = file.read()
        if entry.strip() in content:
            print(f"The entry '{entry.strip()}' already exists in the hosts file.")
            return

    with open(hosts_file, 'a') as file:
        file.write(entry)
        print(f"Added '{entry.strip()}' to the hosts file.")

def mock_server_down():
    mock_server_post("/mock/down", "")

def mock_server_up():
    mock_server_post("/mock/up", "")

def mock_server_get_token():
    return mock_server_get("/mock/token").json().get("token")

def mock_server_get_php_version():
    return mock_server_get("/mock/php_version").json().get("php_version")