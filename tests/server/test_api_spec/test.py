import requests
import time
import sys
from testlib import *

'''
1. Sets up a simple config.
2. Sends one post request.
3. Waits for the heartbeat event and validates the API schema reporting.
'''

def get_api_spec_with_no_body():
    url = "/api/v1/orders?userId=12345&status=pending"
    headers = {
        "Content-Type": "application/json",
    }
    return url, None, headers

def get_api_spec_simple():
    url = "/api/v1/orders?userId=12345&status=pending"
    headers = {
        "Content-Type": "application/json",
        "Authorization": "Bearer your_token_here"
    }
    body = {
        "orderId":"98765",
        "items":[
            {
                "itemId":"abc123",
                "quantity":2,
                "price":29.99,
                "details":{
                    "color":"blue",
                    "size":"M"
                }
            },
            {
                "itemId":"def456",
                "quantity":1,
                "price":19.99,
                "details":{
                    "color":"red",
                    "size":"L"
                }
            }
        ],
        "shippingAddress":{
            "name":"John Doe",
            "street":"1234 Elm St",
            "city":"Some City",
            "state":"CA",
            "zip":"90210",
            "country":"USA"
        },
        "paymentMethod":{
            "type":"credit_card",
            "provider":"Visa",
            "cardNumber":"4111111111111111",
            "expiryDate":"12/25"
        },
        "total":79.97
    }
    return url, body, headers


def get_api_spec_merge():
    url = "/api/v1/orders?userId=12345&status=pending&orderId=80"
    headers = {
        "Content-Type": "application/json",
        "X-API-Key": "abcdef12345"
    }
    body = {
        "orderPlaced": True
    }
    return url, body, headers


def post_form_size_cases():
    base = f"http://localhost:{get_php_port()}"
    limit = 1024 * 1024
    field = b"field="
    response = requests.post(base + "/form-boundary", data=field + b"x" * (limit - len(field)),
                             headers={"Content-Type": "application/x-www-form-urlencoded"})
    assert_response_code_is(response, 200)

    response = requests.post(base + "/file-upload?source=form", data={"field": "small"},
                             files={"file": ("report.txt", b"x" * limit)})
    assert_response_code_is(response, 200)


def run_api_spec_tests(fns, expected_json):
    for fn in fns:
        response = php_server_post(*fn())
        assert_response_code_is(response, 200)
    
    post_form_size_cases()
    mock_server_wait_for_new_events(70)
    
    events = mock_server_get_events()
    assert_events_length_is(events, 2)
    assert_started_event_is_valid(events[0])
    
    assert_event_contains_subset_file(events[1], expected_json)
    schemas = {route["path"]: route["apispec"] for route in events[1]["routes"]}
    assert "field" in schemas["/form-boundary"]["body"]["schema"]["properties"]
    assert schemas["/file-upload"] is None

def run_test():
    run_api_spec_tests([
        get_api_spec_with_no_body,
        get_api_spec_simple,
        get_api_spec_merge,
    ], "expect_api_spec.json")
        
if __name__ == "__main__":
    load_test_args()
    run_test()
