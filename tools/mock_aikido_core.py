from flask import Flask, request, jsonify, Response
import sys
import os
import json
import time
import gzip

app = Flask(__name__)

responses = {
    "config": {},
    "configUpdatedAt": {},
    "lists": {}
}

events = []
server_down = False

excluded_routes = ['mock_get_events', 'mock_tests_simple', 'mock_down', 'mock_up']

def load_config(j):
    configUpdatedAt = int(time.time())
    responses["lists"] = { "success": True, "serviceId": j["serviceId"] }
    if "blockedIPAddresses" in j:
        responses["lists"]["blockedIPAddresses"] = j["blockedIPAddresses"]
        del j["blockedIPAddresses"]
    if "monitoredIPAddresses" in j:
        responses["lists"]["monitoredIPAddresses"] = j["monitoredIPAddresses"]
        del j["monitoredIPAddresses"]
    if "monitoredUserAgents" in j:
        responses["lists"]["monitoredUserAgents"] = j["monitoredUserAgents"]
        del j["monitoredUserAgents"]
    if "blockedUserAgents" in j:
        responses["lists"]["blockedUserAgents"] = j["blockedUserAgents"]
        del j["blockedUserAgents"]
    if "lists_allowedIPAddresses" in j:
        responses["lists"]["allowedIPAddresses"] = j["lists_allowedIPAddresses"]
        del j["lists_allowedIPAddresses"]
    if "userAgentDetails" in j:
        responses["lists"]["userAgentDetails"] = j["userAgentDetails"]
        del j["userAgentDetails"]
    responses["config"] = j
    responses["config"]["configUpdatedAt"] = configUpdatedAt
    responses["configUpdatedAt"] = { "serviceId": 1, "configUpdatedAt": configUpdatedAt }
    print(f"Loaded new runtime config!")

def gzip_response(data):
    json_str = json.dumps(data)
    gzipped = gzip.compress(json_str.encode('utf-8'))
    response = Response(gzipped)
    response.headers['Content-Encoding'] = 'gzip'
    response.headers['Content-Type'] = 'application/json'
    return response

token = None

@app.before_request
def check_server_status():
    global server_down
    if request.endpoint in excluded_routes:
        return None
    if request.headers.get('Authorization') is not None:
        global token
        token = request.headers.get('Authorization')
        print("Token: ", token)
    if server_down:
        return gzip_response({"error": "Service Unavailable"}), 503

@app.route('/config', methods=['GET'])
def get_config():
    return gzip_response(responses["configUpdatedAt"])

@app.route('/api/runtime/config', methods=['GET'])
def get_runtime_config():
    return gzip_response(responses["config"])

@app.route('/api/runtime/firewall/lists', methods=['GET'])
def get_lists_config():
    accept_encoding = request.headers.get('Accept-Encoding', '').lower()
    if 'gzip' not in accept_encoding:
        return jsonify({
            "success": False,
            "error": "Accept-Encoding header must include 'gzip' for firewall lists endpoint"
        }), 400

    return gzip_response(responses["lists"])

@app.route('/api/runtime/events', methods=['POST'])
def post_events():
    print("Got event: ", request.get_json())
    if request.get_json():
        events.append(request.get_json())
    return gzip_response(responses["config"])

@app.route('/mock/config', methods=['POST'])
def mock_set_config():
    load_config(request.get_json())
    return gzip_response({})

@app.route('/mock/down', methods=['POST'])
def mock_down():
    global server_down
    server_down = True
    return gzip_response({})

@app.route('/mock/up', methods=['POST'])
def mock_up():
    global server_down
    server_down = False
    return gzip_response({})

@app.route('/mock/events', methods=['GET'])
def mock_get_events():
    return gzip_response(events)

@app.route('/tests/simple', methods=['GET'])
def mock_tests_simple():
    time.sleep(1)
    return gzip_response("{}")

@app.route('/mock/token', methods=['GET'])
def mock_get_token():
    return gzip_response({"token": token})

if __name__ == '__main__':
    if len(sys.argv) < 2 or len(sys.argv) > 3:
        print("Usage: python mock_server.py <port> [config_file]")
        sys.exit(1)
    
    port = int(sys.argv[1])
    
    if len(sys.argv) == 3:
        config_file = sys.argv[2]
        if os.path.exists(config_file):
            try:
                with open(config_file, 'r') as file:
                    load_config(json.load(file))
            except json.JSONDecodeError:
                print(f"Error: Could not decode JSON from {config_file}")
                sys.exit(1)
        else:
            print(f"Error: File {config_file} not found")
            sys.exit(1)
    
    app.run(host='127.0.0.1', port=port)
