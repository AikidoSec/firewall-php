import requests
import time
import sys
from testlib import *


def generate_users():
    generated_users = []
    for i in range(3500):
        user = f"user{i}"
        generated_users.append(user)
    return generated_users


def run_test():
    generated_users = generate_users()
    for user in generated_users:
        response = php_server_get(f"/?user_id={user}&username={user}")
        assert_response_code_is(response, 200)
    
    mock_server_wait_for_new_events(70)
    
    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_started_event_is_valid(events[0])

    users = [u["id"] for u in events[1]["users"]]
    assert len(users) == 2000, f"Expected 2000 users, got {len(users)}"
    assert generated_users[0] not in users, f"User {generated_users[0]} should not be in reported users"
    assert generated_users[-1] in users, f"User {generated_users[-1]} should be in reported users"

    
if __name__ == "__main__":
    load_test_args()
    run_test()
