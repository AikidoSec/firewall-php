import requests
import time
import sys
from testlib import *


def generate_users():
    generated_users = set()
    while len(generated_users) < 3500:
        user_len = random.randint(1, 20)  # Random route length between 1-20 segments
        user = ''.join(random.choices('abcdefghijklmnopqrstuvwxyz0123456789-_', k=user_len))
        generated_users.add(user)
    return list(generated_users)


def run_test():
    for user in generate_users():
        response = php_server_get(f"/?user_id={user}&username={user}")
        assert_response_code_is(response, 200)
    
    mock_server_wait_for_new_events(70)
    
    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_started_event_is_valid(events[0])
    assert len(events[1]["users"]) == 2000, f"Expected 2000 users, got {len(events[1]['users'])}"

    
if __name__ == "__main__":
    load_test_args()
    run_test()
