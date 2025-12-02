import os
import subprocess
import time
import urllib.request

frankenphp_bin = "frankenphp"
caddyfile_path = "/tmp/frankenphp_test.caddyfile"
log_dir = "/var/log/frankenphp"

caddyfile_base_template = """{{
    frankenphp {{
        num_threads {num_threads}
        max_threads {max_threads}
    }}
}}
"""

site_block_template = """http://:{port} {{
    root * {test_dir}
    php_server {{
{env_vars}
    }}
}}
"""

def create_folder(folder_path):
    if not os.path.exists(folder_path):
        os.makedirs(folder_path)

def frankenphp_create_site_block(test_data):
    env_vars = ""
    for key, value in test_data["env"].items():
        env_vars += f"        env {key} \"{value}\"\n"
    
    return site_block_template.format(
        port=test_data["server_port"],
        test_dir=test_data["test_dir"],
        env_vars=env_vars
    )

def frankenphp_classic_init(tests_dir):
    if os.path.exists(caddyfile_path):
        os.remove(caddyfile_path)
    create_folder(log_dir)
    create_folder('/etc/frankenphp/php.d')

def frankenphp_classic_process_test(test_data):
    test_data["site_block"] = frankenphp_create_site_block(test_data)
    return test_data

def frankenphp_classic_pre_tests(tests_data):
    subprocess.run(['pkill', '-9', '-x', 'frankenphp'], stderr=subprocess.DEVNULL)
    subprocess.run(['pkill', '-9', '-f', 'mock_aikido_core'], stderr=subprocess.DEVNULL)
    subprocess.run(['rm', '-rf', f'{log_dir}/*'])
    subprocess.run(['rm', '-rf', f'/var/log/aikido-*/*'])

    total_workers = len(tests_data)
    threads = total_workers * 2
    
    with open(caddyfile_path, 'w') as f:
        f.write(caddyfile_base_template.format(num_threads=threads, max_threads=threads))
        for test_data in tests_data:
            f.write("\n" + test_data["site_block"])
    
    subprocess.Popen(
        [frankenphp_bin, 'run', '--config', caddyfile_path],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL
    )
    time.sleep(20)
    
    result = subprocess.run(['pgrep', '-x', 'frankenphp'], capture_output=True, text=True)
    if not result.stdout.strip():
        raise RuntimeError("FrankenPHP classic failed to start!")

def frankenphp_classic_start_server(test_data, test_lib_dir, valgrind):
    return None

def frankenphp_classic_uninit():
    subprocess.run(['pkill', '-9', '-x', 'frankenphp'], stderr=subprocess.DEVNULL)
    if os.path.exists(caddyfile_path):
        os.remove(caddyfile_path)

