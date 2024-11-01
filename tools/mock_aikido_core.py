from flask import Flask, request, jsonify
import sys
import os
import json
import time

app = Flask(__name__)

responses = {
    "config": {},
    "configUpdatedAt": {},
}

events = []
server_down = False

excluded_routes = ['mock_get_events', 'mock_tests_simple', 'mock_down', 'mock_up']

@app.before_request
def check_server_status():
    global server_down
    if request.endpoint in excluded_routes:
        return None
    if server_down:
        return jsonify({"error": "Service Unavailable"}), 503

@app.route('/config', methods=['GET'])
def get_config():
    return jsonify(responses["configUpdatedAt"])


@app.route('/api/runtime/config', methods=['GET'])
def get_runtime_config():
    return jsonify(responses["config"])


@app.route('/api/runtime/events', methods=['POST'])
def post_events():
    print("Got event: ", request.get_json())
    if request.get_json():
        events.append(request.get_json())
    return jsonify(responses["config"])


@app.route('/mock/config', methods=['POST'])
def mock_set_config():
    configUpdatedAt = int(time.time())
    responses["config"] = request.get_json()
    responses["config"]["configUpdatedAt"] = configUpdatedAt
    responses["configUpdatedAt"] = { "serviceId": 1, "configUpdatedAt": configUpdatedAt }
    return jsonify({})

@app.route('/mock/down', methods=['POST'])
def mock_down():
    global server_down
    server_down = True
    return jsonify({})

@app.route('/mock/up', methods=['POST'])
def mock_up():
    global server_down
    server_down = False
    return jsonify({})

@app.route('/mock/events', methods=['GET'])
def mock_get_events():
    return jsonify(events)

@app.route('/tests/simple', methods=['GET'])
def mock_tests_simple():
    time.sleep(1)
    return jsonify("{}")

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
                    configUpdatedAt = int(time.time())
                    responses["config"] = json.load(file)
                    responses["config"]["configUpdatedAt"] = configUpdatedAt
                    responses["configUpdatedAt"] = { "serviceId": 1, "configUpdatedAt": configUpdatedAt }
                    print(f"Loaded runtime config from {config_file}")
            except json.JSONDecodeError:
                print(f"Error: Could not decode JSON from {config_file}")
                sys.exit(1)
        else:
            print(f"Error: File {config_file} not found")
            sys.exit(1)
    
    app.run(host='127.0.0.1', port=port)
