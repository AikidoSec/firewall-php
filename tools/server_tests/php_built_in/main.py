import os
import subprocess

def php_built_in_start_server(test_data, test_lib_dir, valgrind):
    server_port = test_data["server_port"]
    
    php_server_process_cmd = ['php', '-S', f'127.0.0.1:{server_port}', '-t', test_data["test_dir"]]
    if valgrind:
        php_server_process_cmd = ['valgrind', f'--suppressions={test_lib_dir}/valgrind.supp', '--track-origins=yes'] + php_server_process_cmd
        
    return subprocess.Popen(
        php_server_process_cmd,
        env=dict(os.environ, **test_data["env"])
    )
