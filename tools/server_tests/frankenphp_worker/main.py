import os
import subprocess
import time

frankenphp_bin = "frankenphp"
caddyfile_path = "/tmp/frankenphp_worker_test.caddyfile"
log_dir = "/var/log/frankenphp"
worker_scripts_dir = "/tmp/frankenphp_workers"

num_workers = 2

def get_php_version():
    """Get PHP version as a tuple (major, minor)"""
    try:
        result = subprocess.run(['php', '-r', 'echo PHP_MAJOR_VERSION.".".PHP_MINOR_VERSION;'], 
                              capture_output=True, text=True, check=True)
        version_str = result.stdout.strip()
        major, minor = version_str.split('.')
        return (int(major), int(minor))
    except:
        return (8, 3)  # Default to newer version

def get_caddyfile_base_template():
    """Get the appropriate caddyfile template based on PHP version"""
    php_version = get_php_version()
    
    # FrankenPHP 1.1.0 (PHP 8.2) doesn't support the {{ global options block
    if php_version == (8, 2):
        # Use a simpler format without global options block
        return ""
    else:
        # Newer versions support the global options block
        return """{{
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
        worker {{
            file {worker_script}
            num {num_workers}
        }}
    }}
}}
"""

worker_script_template = """<?php
$test_dir = '{test_dir}';

$handler = function() use ($test_dir) {{
    $uri = $_SERVER['REQUEST_URI'] ?? '/';
    $path = parse_url($uri, PHP_URL_PATH) ?: '/';
    
    if ($path === '/' || $path === '') {{
        $file = $test_dir . '/index.php';
    }} else {{
        $file = $test_dir . $path;
        if (!file_exists($file) || !is_file($file)) {{
            $file = $test_dir . '/index.php';
        }}
    }}
    
    if (file_exists($file) && is_file($file)) {{
        include $file;
    }} else {{
        http_response_code(404);
        echo "Not Found";
    }}
}};

for ($nbWorkers = frankenphp_handle_request($handler); $nbWorkers > 0; $nbWorkers = frankenphp_handle_request($handler)) {{
    gc_collect_cycles();
}}
"""

def create_folder(folder_path):
    if not os.path.exists(folder_path):
        os.makedirs(folder_path)

def frankenphp_worker_create_script(test_dir, test_name):
    worker_script_path = os.path.join(worker_scripts_dir, f"{test_name}.php")
    worker_script_content = worker_script_template.format(test_dir=test_dir)
    
    with open(worker_script_path, 'w') as f:
        f.write(worker_script_content)
    
    return worker_script_path

def frankenphp_worker_create_site_block(test_data, worker_script_path):
    env_vars = f"        env DOCUMENT_ROOT \"{test_data['test_dir']}\"\n"
    for key, value in test_data["env"].items():
        env_vars += f"        env {key} \"{value}\"\n"
    
    return site_block_template.format(
        port=test_data["server_port"],
        test_dir=test_data["test_dir"],
        worker_script=worker_script_path,
        env_vars=env_vars,
        num_workers=num_workers
    )

def frankenphp_worker_init(tests_dir):
    if os.path.exists(caddyfile_path):
        os.remove(caddyfile_path)
    subprocess.run(['rm', '-rf', f'{worker_scripts_dir}/*'])
    create_folder(log_dir)
    create_folder('/etc/frankenphp/php.d')
    create_folder(worker_scripts_dir)

def frankenphp_worker_process_test(test_data):
    test_name = test_data["test_name"]
    worker_script_path = frankenphp_worker_create_script(test_data["test_dir"], test_name)
    test_data["site_block"] = frankenphp_worker_create_site_block(test_data, worker_script_path)
    return test_data

def frankenphp_worker_pre_tests(tests_data):
    subprocess.run(['pkill', '-9', '-x', 'frankenphp'], stderr=subprocess.DEVNULL)
    subprocess.run(['pkill', '-9', '-f', 'mock_aikido_core'], stderr=subprocess.DEVNULL)
    subprocess.run(['rm', '-rf', f'{log_dir}/*'])
    subprocess.run(['rm', '-rf', f'/var/log/aikido-*/*'])
    subprocess.run(['rm', '-rf', f'{worker_scripts_dir}/*'])
    
    total_workers = len(tests_data)
    threads = total_workers * 3

    with open(caddyfile_path, 'w') as f:
        base_template = get_caddyfile_base_template()
        if base_template:
            f.write(base_template.format(num_threads=threads, max_threads=threads*2))
        for test_data in tests_data:
            f.write("\n" + test_data["site_block"])
    
    print(f"Caddyfile prepared for {len(tests_data)} tests with {threads} threads")
    return threads

def frankenphp_worker_start_server(test_data, test_lib_dir, valgrind):
    result = subprocess.run(['pgrep', '-x', 'frankenphp'], capture_output=True, text=True)
    
    if not result.stdout.strip():
        print("Starting FrankenPHP worker server...")
        process = subprocess.Popen(
            [frankenphp_bin, 'run', '--config', caddyfile_path]
        )
        time.sleep(2)
        
        result = subprocess.run(['pgrep', '-x', 'frankenphp'], capture_output=True, text=True)
        if not result.stdout.strip():
            raise RuntimeError("FrankenPHP worker failed to spawn!")
        
        print("FrankenPHP worker process started")
    else:
        print("FrankenPHP worker is already running")
    
    return None

def frankenphp_worker_uninit():
    subprocess.run(['pkill', '-9', '-x', 'frankenphp'], stderr=subprocess.DEVNULL)
    if os.path.exists(caddyfile_path):
        os.remove(caddyfile_path)
    subprocess.run(['rm', '-rf', f'{worker_scripts_dir}/*'])

