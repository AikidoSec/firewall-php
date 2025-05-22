import requests
import time
import sys
from testlib import *


def generate_domains():
    generated_domains = set()
    while len(generated_domains) < 2100:
        domain_len = random.randint(20, 30)  # Random route length between 1-20 segments
        domain = ''.join(random.choices('abcdefghijklmnopqrstuvwxyz0123456789-_', k=domain_len))
        generated_domains.add(domain)
    return list(generated_domains)


def run_test():
    generated_domains = generate_domains()
    for domain in generated_domains:
        response = php_server_get(f"/?domain={domain}")
        assert_response_code_is(response, 200)
    
    mock_server_wait_for_new_events(310)
    
    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_started_event_is_valid(events[0])

    domains = [d["hostname"] for d in events[1]["hostnames"]]
    assert len(domains) == 2000, f"Expected 2000 domains, got {len(domains)}"
    assert generated_domains[0] + ".com" not in domains, f"Domain {generated_domains[0]} should not be in reported domains"
    assert generated_domains[-1] + ".com" in domains, f"Domain {generated_domains[-1]} should be in reported domains"

    
if __name__ == "__main__":
    load_test_args()
    run_test()
