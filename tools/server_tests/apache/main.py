import os
import subprocess
import re
import pwd
import grp
import psutil
import time

if os.path.exists('/etc/httpd'):
    # Centos
    apache_binary = "httpd"
    apache_server_root = "/etc/httpd"
    apache_conf_global_file = f"{apache_server_root}/conf/httpd.conf"
    apache_conf_proxy_module_file = f"{apache_server_root}/conf.modules.d/00-proxy.conf"
    apache_conf_proxy_h2_module_file = f"{apache_server_root}/conf.modules.d/10-proxy_h2.conf"
    apache_conf_mpm_worker_file = f"{apache_server_root}/conf.modules.d/00-mpm.conf"
    apache_conf_mpm_event_file = f"{apache_server_root}/conf.modules.d/00-mpm.conf"
    apache_conf_mpm_prefork_file = f"{apache_server_root}/conf.modules.d/00-mpm.conf"
    apache_conf_folder = f"{apache_server_root}/conf.d"
    apache_log_folder = "/var/log/httpd"
    apache_run_folder = "/run/httpd"
    apache_include_conf = """Include conf.modules.d/*.conf
IncludeOptional conf.d/*.conf"""
    apache_error_log = "logs/error_log"
else:
    # Debian
    apache_binary = "apache2"
    apache_server_root = "/etc/apache2"
    apache_conf_global_file = f"{apache_server_root}/apache2.conf"
    apache_conf_proxy_module_file = f"{apache_server_root}/mods-available/proxy.conf"
    apache_conf_proxy_h2_module_file = f"{apache_server_root}/mods-available/proxy_http2.load"
    apache_conf_mpm_worker_file = f"{apache_server_root}/mods-available/mpm_worker.conf"
    apache_conf_mpm_event_file = f"{apache_server_root}/mods-available/mpm_event.conf"
    apache_conf_mpm_prefork_file = f"{apache_server_root}/mods-available/mpm_prefork.conf"
    apache_conf_folder = f"{apache_server_root}/sites-available"
    apache_log_folder = "/var/log/apache2"
    apache_run_folder = "/run/apache2"
    apache_include_conf = """IncludeOptional mods-enabled/*.load
IncludeOptional mods-enabled/*.conf"""
    apache_error_log = f"{apache_log_folder}/error.log"

apache_conf_template = """
ServerRoot "{server_root}"
ServerName "localhost"
PidFile {server_run}/{server_binary}-{name}.pid

User {user}
Group {user}
ServerAdmin root@localhost
Listen {port}

ErrorLog {error_log}

LogFormat "%h %l %u %t %r %>s %b" combined

{optional_conf}

<IfModule mime_module>
    TypesConfig /etc/mime.types
    AddType application/x-compress .Z
    AddType application/x-gzip .gz .tgz
    AddType text/html .shtml
    AddOutputFilter INCLUDES .shtml
</IfModule>

<IfModule mpm_prefork_module>
    StartServers        3
    MinSpareServers     3
    MaxSpareServers     3
</IfModule>

<VirtualHost *:{port}>
    ServerName localhost
    DocumentRoot {test_dir}
    DirectoryIndex index.php

    <Directory {test_dir}>
        Options Indexes FollowSymLinks
        AllowOverride All
        Require all granted
        
        RewriteEngine On
        RewriteCond %{{REQUEST_FILENAME}} !-f
        RewriteCond %{{REQUEST_FILENAME}} !-d
        RewriteRule ^(.*)$ index.php [L]
        
        SetEnvIf Authorization "(.*)" HTTP_AUTHORIZATION=$1
    </Directory>

    ErrorLog {log_dir}/error_{name}.log
    CustomLog {log_dir}/access_{name}.log combined

    php_admin_value error_reporting "E_ALL"
    php_admin_flag  log_errors On
    php_admin_flag  display_errors On
    php_admin_value error_log "/proc/self/fd/2"
    LogLevel debug

</VirtualHost>
"""

def append_if_not_exists(file_path, content):
    try:
        # Open the file in read mode to check for existing content
        with open(file_path, 'r') as file:
            existing_content = file.read()

        # Check if the content already exists in the file
        if content not in existing_content:
            # If the content does not exist, append it
            with open(file_path, 'a') as file:
                file.write(content)
            print(f"Content appended to {file_path}.")
        else:
            print("Content already exists in the file, no action taken.")
    except FileNotFoundError:
        # If the file does not exist, create it and write the content
        with open(file_path, 'w') as file:
            file.write(content)
        print(f"File created and content added to {file_path}.")


def modify_apache_conf(file_path):
    try:
        with open(file_path, 'r') as file:
            content = file.read()

        content = content.replace('User www-data', 'User root')
        content = content.replace('Group www-data', 'Group root')
        content = content.replace('User apache', 'User root')
        content = content.replace('Group apache', 'Group root')

        with open(file_path, 'w') as file:
            file.write(content)

        print(f"nginx.conf has been updated to use 'user root;'.")
    except FileNotFoundError:
        print(f"Error: File {file_path} not found.")
    except Exception as e:
        print(f"Error: {e}")


def toggle_config_line(file_path, line_to_check, comment_ch, enable=False):
    if not os.path.exists(file_path):
        print(f"File '{file_path}' does not exist.")
        return
    
    with open(file_path, 'r') as file:
        lines = file.readlines()

    commented_line_pattern = r"\s*" + re.escape(line_to_check.strip()) + r".*"

    if enable:
        commented_line_pattern = "\s*" + comment_ch + commented_line_pattern

    # Initialize a flag to track changes
    changes_made = False

    # Iterate through the lines and check for the commented line
    with open(file_path, 'w') as file:
        for line in lines:
            if re.match(commented_line_pattern, line):
                if enable:
                    # Uncomment it if enable is True
                    file.write(line.replace(comment_ch, "", 1).lstrip())
                else:
                    # Comment it if enable is False
                    file.write(f"{comment_ch} {line}".lstrip())

                changes_made = True
            else:
                # Otherwise, write the line as-is
                file.write(line)

    if changes_made:
        if enable:
            print(f"The line '{line_to_check}' was uncommented.")
        else:
            print(f"The line '{line_to_check}' was commented.")


apache_user = None
prev_owning_user = ""
prev_owning_group = ""

def select_apache_user():
    global apache_user
    users = pwd.getpwall()
    usernames = [user.pw_name for user in users]
    print("Users on system: ", usernames)
    for u in ["apache", "www-data"]:
        if u in usernames:
            apache_user = u
            break
    
    assert apache_user is not None
        
    print("Selected apache user: ", apache_user)


def get_user_and_group(folder_path):
    # Get the folder's status, which includes owner and group info
    folder_stat = os.stat(folder_path)

    # Get the user ID and group ID
    user_id = folder_stat.st_uid
    group_id = folder_stat.st_gid

    # Get the username from the user ID
    user_name = pwd.getpwuid(user_id).pw_name

    # Get the group name from the group ID
    group_name = grp.getgrgid(group_id).gr_name
    return user_name, group_name


def apache_create_config_file(test_name, test_dir, server_port, env):
    apache_config = apache_conf_template.format(
        server_root = apache_server_root,
        server_binary = apache_binary,
        server_run = apache_run_folder,
        name = test_name,
        port = server_port,
        test_dir = test_dir,
        log_dir = apache_log_folder,
        user = apache_user,
        optional_conf = apache_include_conf,
        error_log = apache_error_log
    )
    
    apache_config_file = os.path.join(test_dir, f"{test_name}.conf")
    with open(apache_config_file, "w") as f:
        f.write(apache_config)

    print(f"Configured apache config for {test_name}")
    return apache_config_file


def add_user_group_access(full_path, user, group):
    try:
        # Split the full path into individual directories
        path_parts = full_path.split(os.sep)

        # Traverse through each part of the path and apply permissions
        for i in range(1, len(path_parts) + 1):
            current_path = os.sep.join(path_parts[:i])
            if current_path:  # Avoid empty strings for the root "/"
                # print(f"Setting permissions for {current_path}")
                
                # Change ownership of the directory
                subprocess.run(['chown', f'{user}:{group}', current_path], check=True)

                # Ensure the execute permission (search permission) on directories
                subprocess.run(['chmod', '775', current_path], check=True)
        
        print(f"Successfully added access to full path '{full_path}' for user '{user}' and group '{group}'.")
    except subprocess.CalledProcessError as e:
        print(f"Failed to modify permissions: {e}")


def apache_mod_php_init(tests_dir):
    select_apache_user()

    subprocess.run(['pkill', apache_binary])
    subprocess.run(['rm', '-rf', f'{apache_log_folder}/*'])
    subprocess.run(['mkdir', '-p', apache_log_folder], check=True)
    subprocess.run(['chown', f'{apache_user}:{apache_user}', apache_log_folder], check=True)
    subprocess.run(['chmod', '755', apache_log_folder], check=True)
    
    
    toggle_config_line(apache_conf_proxy_module_file, "LoadModule proxy_fcgi_module", "#")
    toggle_config_line(apache_conf_proxy_h2_module_file, "LoadModule proxy_http2_module", "#")
    
    toggle_config_line(apache_conf_mpm_worker_file, "LoadModule mpm_worker_module", "#")
    toggle_config_line(apache_conf_mpm_event_file, "LoadModule mpm_event_module", "#")
    toggle_config_line(apache_conf_mpm_prefork_file, "LoadModule mpm_prefork_module", "#", enable=True)
    
    global prev_owning_user, prev_owning_group
    prev_owning_user, prev_owning_group = get_user_and_group(tests_dir)
    print(f"Got previous owning user:group -> {prev_owning_user}:{prev_owning_group}")
    

def apache_mod_php_process_test(test_data):
    test_dir = test_data["test_dir"]
    server_port = test_data["server_port"]
    test_data["apache_config"] = apache_create_config_file(test_data["test_name"], test_dir, server_port, test_data["env"])
    
    global apache_user
    add_user_group_access(os.path.join(test_dir, "index.php"), apache_user, apache_user)
    
    # append_if_not_exists(apache_conf_global_file, f"Listen {server_port}\n")
    return test_data


def apache_mod_php_pre_tests():
    pass


def apache_mod_php_start_server(test_data, test_lib_dir, valgrind):
    print([f'/usr/sbin/{apache_binary}', '-f', test_data["apache_config"]])
    return subprocess.Popen([f'/usr/sbin/{apache_binary}', '-f', test_data["apache_config"]], env=test_data["env"])


def apache_mod_php_uninit():
    subprocess.run(['pkill', apache_binary])
    subprocess.run(['chown', '-R', f'{prev_owning_user}:{prev_owning_group}', '../'])
