import requests
import time
import sys
from testlib import *

'''
1. Sets AIKIDO_DISABLE to 1.
2. Sends one post request that should generate a path traversal detection.
3. Checks that the file indeed was opened and path traversal was not blocked.
4. Checks that the started event is not submitted.
5. Checks that no events are submitted event after 1 minute.
'''

def run_test():
    response = php_server_post("/testDetection", {"folder": "../../../.."})
    assert_response_code_is(response, 200)
    assert_response_body_contains(response, "File opened!")
    
    if mock_server_get_platform_name() != "apache2handler":
        events = mock_server_get_events()
        assert_events_length_is(events, 0)
        
        time.sleep(65)
        
        events = mock_server_get_events()
        assert_events_length_is(events, 0)
        
        
if __name__ == "__main__":
    load_test_args()
    run_test()
