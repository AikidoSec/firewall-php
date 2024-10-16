import os
import subprocess
import re
import pwd
import psutil
import time

nginx_config_dir = "/etc/nginx/conf.d"

users = pwd.getpwall()
usernames = [user.pw_name for user in users]
print("Users on system: ", usernames)

def get_user_of_process(process_name):
    # Iterate over all running processes
    for proc in psutil.process_iter(['pid', 'name', 'username']):
        try:
            # Check if the process name matches
            if proc.info['name'] == process_name:
                print(f"Process '{process_name}' (PID: {proc.info['pid']}) is running under user: {proc.info['username']}")
        except (psutil.NoSuchProcess, psutil.AccessDenied, psutil.ZombieProcess):
            pass

nginx_conf_template = """
server {{
    listen {port};
    server_name {name};

    root {test_dir};
    index index.php;

    location / {{
        try_files $uri $uri/ /index.php?$args;
    }}

    location ~ \.php$ {{
        fastcgi_split_path_info ^(.+\.php)(/.+)$;
        fastcgi_pass unix:/run/php-fpm/php-fpm-{name}.sock;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_index index.php;
        include fastcgi.conf;
    }}
}}
"""

php_fpm_conf_template = """[{name}]
user = {user}
group = {user}
listen = /run/php-fpm/php-fpm-{name}.sock
listen.owner = {user}
listen.group = {user}
pm = dynamic
pm.max_children = 5
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 3
clear_env = no
catch_workers_output = yes
access.log = /var/log/php-fpm/access-{name}.log

php_admin_value[error_log] = /var/log/php-fpm/error-{name}.log
php_admin_flag[log_errors] = on
"""

def create_folder(folder_path):
    if not os.path.exists(folder_path):
        os.makedirs(folder_path)

def enable_config_line(file_path, line_to_check, comment_ch):
    # Read the nginx.conf file
    with open(file_path, 'r') as file:
        lines = file.readlines()

    # Prepare a regex pattern to match the commented line
    commented_line_pattern = r"\s*" + comment_ch + r"\s*" + re.escape(line_to_check.strip()) + r"\s*"

    # Initialize a flag to track changes
    changes_made = False

    # Iterate through the lines and check for the commented line
    with open(file_path, 'w') as file:
        for line in lines:
            if re.match(commented_line_pattern, line):
                # If the line is commented, uncomment it
                file.write(line.replace(comment_ch, "", 1).lstrip())
                changes_made = True
            else:
                # Otherwise, write the line as-is
                file.write(line)

    if changes_made:
        print(f"The line '{line_to_check}' was uncommented.")


def modify_nginx_conf(file_path):
    try:
        # Read the nginx configuration file
        with open(file_path, 'r') as file:
            content = file.read()

        # Replace 'user nginx;' or 'user www-data;' with 'user root;'
        content = content.replace('user nginx;', 'user root;')
        content = content.replace('user www-data;', 'user root;')

        # Write the modified content back to the file
        with open(file_path, 'w') as file:
            file.write(content)

        print(f"nginx.conf has been updated to use 'user root;'.")
    except FileNotFoundError:
        print(f"Error: File {file_path} not found.")
    except Exception as e:
        print(f"Error: {e}")

def nginx_create_conf_file(test_name, test_dir, server_port):
    nginx_config = nginx_conf_template.format(
        name = test_name,
        port = server_port,
        test_dir = test_dir
    )

    nginx_config_file = os.path.join(nginx_config_dir, f"{test_name}.conf")
    with open(nginx_config_file, "w") as fpm_file:
        fpm_file.write(nginx_config)

    print(f"Configured nginx config for {test_name}")


def php_fpm_create_conf_file(test_dir, test_name, user):
    nginx_user = "root"
    for u in ["nginx", "www-data"]:
        if u in usernames:
            nginx_user = u
            break
        
    print("Selected nginx user: ", nginx_user)
    php_fpm_config = php_fpm_conf_template.format(
        name = test_name,
        user = user
    )
        
    php_fpm_config_file_path = os.path.join(test_dir, f"{test_name}.conf")
    with open(php_fpm_config_file_path, "w") as fpm_file:
        fpm_file.write(php_fpm_config)

    print(f"Configured PHP-FPM config for {test_name}")
    
    return php_fpm_config_file_path


def prepare_nginx_php_fpm(test_data):
    enable_config_line("/etc/nginx/nginx.conf", "include /etc/nginx/conf.d/*.conf;", '#')
    nginx_create_conf_file(test_data["test_name"], test_data["test_dir"], test_data["server_port"])

    test_data["fpm_config"] = php_fpm_create_conf_file(test_data["test_dir"], test_data["test_name"], "root")
    return test_data


def pre_nginx_php_fpm():
    subprocess.run(['pkill', 'nginx'])
    subprocess.run(['pkill', 'php-fpm'])
    subprocess.run(['rm', '-rf', '/var/log/nginx/*'])
    subprocess.run(['rm', '-rf', '/var/log/php-fpm/*'])
    subprocess.run(['rm', '-rf', '/var/log/aikido-1.0.79/*'])
    create_folder("/run/php-fpm")
    create_folder("/var/log/php-fpm")
    modify_nginx_conf("/etc/nginx/nginx.conf")
    subprocess.run(['nginx'], check=True)
    print("nginx server restarted!")
    time.sleep(5)


def handle_nginx_php_fpm(test_data, test_lib_dir, valgrind):
    php_fpm_command = ["/usr/sbin/php-fpm", "--force-stderr", "--nodaemonize", "--allow-to-run-as-root", "--fpm-config", test_data["fpm_config"]]
    print("PHP-FPM command: ", php_fpm_command)
    return [subprocess.Popen(php_fpm_command, env=test_data["env"])]


def done_nginx_php_fpm():
    subprocess.run(['pkill', 'nginx'], check=True)