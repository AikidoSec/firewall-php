import os
import threading
import subprocess
import random
import time
import sys
import json
import argparse
import socket
from server_tests.php_built_in.main import php_built_in_start_server
from server_tests.apache.main import apache_mod_php_init, apache_mod_php_process_test, apache_mod_php_pre_tests, apache_mod_php_start_server, apache_mod_php_uninit
from server_tests.nginx.main import nginx_php_fpm_init, nginx_php_fpm_process_test, nginx_php_fpm_pre_tests, nginx_php_fpm_start_server, nginx_php_fpm_uninit

INIT = 0
PROCESS_TEST = 1
PRE_TESTS = 2
START_SERVER = 3
UNINIT = 4
    
servers = {
    "php-built-in": (
        None,
        None,
        None,
        php_built_in_start_server,
        None
    ),
    "apache-mod-php": ( 
        apache_mod_php_init, 
        apache_mod_php_process_test, 
        apache_mod_php_pre_tests, 
        apache_mod_php_start_server, 
        apache_mod_php_uninit
    ),
    "nginx-php-fpm": ( 
        nginx_php_fpm_init, 
        nginx_php_fpm_process_test, 
        nginx_php_fpm_pre_tests, 
        nginx_php_fpm_start_server, 
        nginx_php_fpm_uninit
    ),
}

used_ports = set()
passed_tests = []
failed_tests = []

lock = threading.Lock()

def is_port_in_active_use(port):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        result = sock.connect_ex(('127.0.0.1', port))
        return result == 0

def generate_unique_port():
    with lock:
        while True:
            port = random.randint(1024, 9999)
            if port not in used_ports and not is_port_in_active_use(port):
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


def handle_test_scenario(data, root_tests_dir, test_lib_dir, server, benchmark, valgrind, debug):
    test_name = data["test_name"]
    mock_port = data["mock_port"]
    server_port = data["server_port"]
    try:
        print(f"Running {test_name}...")
        print(f"Starting mock server on port {mock_port} with start_config.json for {test_name}...")
        mock_aikido_core = subprocess.Popen(["python3", "mock_aikido_core.py", str(mock_port), data["config_path"]])
        time.sleep(5)

        print(f"Starting {server} server on port {server_port} for {test_name}...")
        
        server_start = servers[server][START_SERVER]
        server_process = server_start(data, test_lib_dir, valgrind)

        time.sleep(10)

        test_script_name = "test.py"
        test_script_cwd = data["test_dir"]
        if benchmark:
            print(f"Running benchmark for {test_name}...")
            test_script_name = "benchmark.py"
            test_script_cwd = root_tests_dir
        else:
            print(f"Running test.py for {test_name}...")
            
        subprocess.run(["python3", test_script_name, str(server_port), str(mock_port), test_name], 
                       env=dict(os.environ, PYTHONPATH=f"{test_lib_dir}:$PYTHONPATH"),
                       cwd=test_script_cwd,
                       check=True, timeout=600, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        
        passed_tests.append(test_name)

    except subprocess.CalledProcessError as e:
        print(f"Error in testing scenario {test_name}:")
        print(f"Exception output: {e.output}")
        print(f"Test exit code: {e.returncode}")
        print(f"Test stdout: {e.stdout.decode()}")
        print(f"Test stderr: {e.stderr.decode()}")
        failed_tests.append(test_name)
        
    except subprocess.TimeoutExpired:
        print(f"Error in testing scenario {test_name}:")
        print(f"Execution timed out.")
        failed_tests.append(test_name)
        
    finally:
        if server_process:
            server_process.terminate()
            server_process.wait()
            print(f"PHP server on port {server_port} stopped.")

        if mock_aikido_core:
            mock_aikido_core.terminate()
            mock_aikido_core.wait()
            print(f"Mock server on port {mock_port} stopped.")



def main(root_tests_dir, test_lib_dir, test_dirs, server="php-built-in", benchmark=False, valgrind=False, debug=False):
    server_init = servers[server][INIT] 
    if server_init is not None:
        server_init(root_tests_dir)
        
    tests_data = []
    for test_dir in test_dirs:
        mock_port = generate_unique_port()
        test_data = {
            "test_name": os.path.basename(os.path.normpath(test_dir)),
            "test_dir": test_dir,
            "mock_port": mock_port,
            "server_port": generate_unique_port(),
            "config_path": os.path.join(test_dir, "start_config.json"),
            "env_path": os.path.join(test_dir, "env.json")
        }

        env = {
            "AIKIDO_LOG_LEVEL": "DEBUG" if debug else "ERROR",
            "AIKIDO_DISK_LOGS": "1" if debug else "0",
            "AIKIDO_TOKEN": "AIK_TOKEN_TEST_" + str(mock_port),
            "AIKIDO_ENDPOINT": f"http://localhost:{mock_port}/",
            "AIKIDO_REALTIME_ENDPOINT": f"http://localhost:{mock_port}/",
        }
        env.update(load_env_from_json(test_data["env_path"]))
        env = {k: v for k, v in env.items() if v != ""}
        test_data["env"] = env
        
        server_process_test = servers[server][PROCESS_TEST]
        if server_process_test is not None:
            test_data = server_process_test(test_data)
        tests_data.append(test_data)
            
    if servers[server][PRE_TESTS] is not None:
        test_data = servers[server][PRE_TESTS]()
            
    threads = []
    for test_data in tests_data:
        args = (test_data, root_tests_dir, test_lib_dir, server, benchmark, valgrind, debug)
        thread = threading.Thread(target=handle_test_scenario, args=args)
        threads.append(thread)
        thread.start()
        time.sleep(10)

    for thread in threads:
        thread.join()
        
    server_uninit = servers[server][UNINIT]
    if server_uninit is not None:
        server_uninit()



if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Script for running PHP server tests with Aikido Firewall installed.")
    parser.add_argument("root_folder_path", type=str, help="Path to the root folder of the tests to be ran.")
    parser.add_argument("test_lib_dir", type=str, help="Directory for the test libraries.")
    parser.add_argument("--test", type=str, action='append', default=None, help="Run specific tests from the root folder (can be specified multiple times).")
    parser.add_argument("--benchmark", action="store_true", help="Enable benchmarking.")
    parser.add_argument("--valgrind", action="store_true", help="Enable valgrind.")
    parser.add_argument("--debug", action="store_true", help="Enable debugging logs.")
    parser.add_argument("--server", type=str, choices=["php-built-in", "apache-mod-php", "nginx-php-fpm"], default="php-built-in", help="Select the type of server testing.")
    parser.add_argument("--max-tests", type=int, default=0, help="Maximum number of tests to execute.")
    parser.add_argument("--max-runs", type=int, default=1, help="Maximum number of test runs.")

    # Parse arguments
    args = parser.parse_args()

    # Extract values from parsed arguments
    root_folder = os.path.abspath(args.root_folder_path)
    test_lib_dir = os.path.abspath(args.test_lib_dir)

    test_dirs = []
    if args.test:
        test_dirs = [os.path.join(root_folder, t) for t in args.test]
    else:
        test_dirs = [f.path for f in os.scandir(root_folder) if f.is_dir()]

    i = 1
    while i <= args.max_runs:
        random.shuffle(test_dirs)
        if args.max_tests != 0:
            test_dirs = test_dirs[:args.max_tests]

        print(f"Starting test run number {i}...")

        main(root_folder, test_lib_dir, test_dirs, args.server, args.benchmark, args.valgrind, args.debug)

        if len(failed_tests) == 0:
            break

        test_dirs = [os.path.join(root_folder, t) for t in failed_tests]

        i += 1

    print_test_results("Passed tests:", passed_tests)
    print_test_results("Failed tests:", failed_tests)
    assert failed_tests == [], f"Found failed tests: {failed_tests}"
    print("All tests passed!")
