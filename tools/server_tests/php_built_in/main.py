import os
import subprocess

def handle_php_built_in(test_data, test_lib_dir, valgrind):
    server_port = test_data["server_port"]
    
    php_server_process_cmd = ['php', '-S', f'localhost:{server_port}', '-t', test_data["test_dir"]]
    if valgrind:
        php_server_process_cmd = ['valgrind', f'--suppressions={test_lib_dir}/valgrind.supp', '--track-origins=yes'] + php_server_process_cmd
        
    return [subprocess.Popen(
        php_server_process_cmd,
        env=test_data["env"]
    )]
