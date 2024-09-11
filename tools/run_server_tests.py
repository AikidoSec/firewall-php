import os
import threading
import subprocess
import random
import time
import sys
import json
import argparse

used_ports = set()
passed_tests = []
failed_tests = []

def generate_unique_port():
    while True:
        port = random.randint(1024, 65535)
        if port not in used_ports:
            used_ports.add(port)
            return port

def load_env_from_json(file_path):
    if not os.path.exists(file_path):
        return {}

    with open(file_path) as f:
        env_vars = json.load(f)
        return env_vars
    
def print_test_results(s, tests):
    if not len(tests):
        return
    
    print(s)
    for t in tests:
        print(f"\t- {t}")

def handle_test_scenario(root_tests_dir, test_dir, test_lib_dir, benchmark):
    try:
        # Generate unique ports for mock server and PHP server.
        mock_port = generate_unique_port()
        php_port = generate_unique_port()

        test_name = os.path.basename(os.path.normpath(test_dir))

        config_path = os.path.join(test_dir, 'start_config.json')
        env_file_path = os.path.join(test_dir, 'env.json')

        print(f"Running {test_name}...")
        print(f"Starting mock server on port {mock_port} with start_config.json for {test_name}...")
        mock_aikido_core = subprocess.Popen(['python3', 'mock_aikido_core.py', str(mock_port), config_path])
        time.sleep(5)

        print(f"Starting PHP server on port {php_port} for {test_name}...")
        env = os.environ.copy()
        env.update(load_env_from_json(env_file_path))
        env.update({
            'AIKIDO_LOG_LEVEL': 'DEBUG',
            'AIKIDO_TOKEN': 'AIK_RUNTIME_MOCK',
            'AIKIDO_ENDPOINT': f'http://localhost:{mock_port}/',
            'AIKIDO_CONFIG_ENDPOINT': f'http://localhost:{mock_port}/',
        })
        php_server_process = subprocess.Popen(
            ['valgrind', 'php', '-S', f'localhost:{php_port}', '-t', test_dir],
            env=env
        )
        time.sleep(5)

        test_script_name = "test.py"
        test_script_cwd = test_dir
        if benchmark:
            print(f"Running benchmark for {test_name}...")
            test_script_name = "benchmark.py"
            test_script_cwd = root_tests_dir
        else:
            print(f"Running test.py for {test_name}...")
            
        subprocess.run(['python3', test_script_name, str(php_port), str(mock_port)], 
                       env=dict(os.environ, PYTHONPATH=f"{test_lib_dir}:$PYTHONPATH"),
                       cwd=test_script_cwd,
                       check=True, timeout=600)
        
        passed_tests.append(test_name)

    except subprocess.CalledProcessError as e:
        print(f"Error in testing scenario {test_name}:")
        print(f"Test output: {e.output}")
        failed_tests.append(test_name)
        
    except subprocess.TimeoutExpired:
        print(f"Error in testing scenario {test_name}:")
        print(f"Execution timed out.")
        failed_tests.append([test_name, "Timeout"])
        
    finally:
        if php_server_process:
            php_server_process.terminate()
            php_server_process.wait()
            print(f"PHP server on port {php_port} stopped.")

        if mock_aikido_core:
            mock_aikido_core.terminate()
            mock_aikido_core.wait()
            print(f"Mock server on port {mock_port} stopped.")


def main(root_tests_dir, test_lib_dir, specific_test=None, benchmark=False):
    if specific_test:
        specific_test = os.path.join(root_tests_dir, specific_test)
        handle_test_scenario(root_tests_dir, specific_test, test_lib_dir, benchmark)
    else:
        test_dirs = [f.path for f in os.scandir(root_tests_dir) if f.is_dir()]
        threads = []

        for test_dir in test_dirs:
            thread = threading.Thread(target=handle_test_scenario, args=(root_tests_dir, test_dir, test_lib_dir, benchmark))
            threads.append(thread)
            thread.start()

        for thread in threads:
            thread.join()
            
    print_test_results("Passed tests:", passed_tests)
    print_test_results("Failed tests:", failed_tests)
    assert failed_tests == [], f"Found failed tests: {failed_tests}"
    print("All tests passed!")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Script for running PHP server tests with Aikido Firewall installed.")
    parser.add_argument('root_folder_path', type=str, help='Path to the root folder of the tests to be ran.')
    parser.add_argument('test_lib_dir', type=str, help='Directory for the test libraries.')
    parser.add_argument('--test', type=str, default=None, help='Run a single test from the root folder.')
    parser.add_argument('--benchmark', action='store_true', help='Enable benchmarking.')

    # Parse arguments
    args = parser.parse_args()

    # Extract values from parsed arguments
    root_folder = os.path.abspath(args.root_folder_path)
    test_lib_dir = os.path.abspath(args.test_lib_dir)
    main(root_folder, test_lib_dir, args.test, args.benchmark)