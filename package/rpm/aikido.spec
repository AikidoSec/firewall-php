Name:           aikido-php-firewall
Version:        1.3.5
Release:        1
Summary:        Aikido PHP Extension
License:        GPL
URL:            https://aikido.dev
Obsoletes:      %{name} < %{version}
Provides:       %{name} = %{version}
Source0:        %{name}-%{version}.tar.gz

%description
Aikido PHP extension and agent.

%prep
%setup -q

%install
mkdir -p %{buildroot}/opt/aikido-%{version}
cp -rf opt/aikido-%{version}/* %{buildroot}/opt/aikido-%{version}

%post
#!/bin/bash

echo "Starting the installation process for Aikido PHP Firewall v%{version}..."

mkdir -p /var/log/aikido-%{version}
chmod 777 /var/log/aikido-%{version}

# Find all PHP versions installed
PHP_VERSIONS=()
if command -v php -v >/dev/null 2>&1; then
    # Add system default PHP version
    PHP_VERSIONS+=($(php -v | grep -oP 'PHP \K\d+\.\d+' | head -n 1))
fi

# Check common PHP installation paths
for php_path in /usr/bin/php* /usr/local/bin/php*; do
    if [[ -x "$php_path" && "$php_path" =~ php([0-9]+\.[0-9]+)$ ]]; then
        version=$("$php_path" -v | grep -oP 'PHP \K\d+\.\d+' | head -n 1)
        if [[ ! " ${PHP_VERSIONS[@]} " =~ " ${version} " ]]; then
            PHP_VERSIONS+=("$version")
        fi
    fi
done

if [ ${#PHP_VERSIONS[@]} -eq 0 ]; then
    echo "No PHP versions found! Exiting!"
    exit 1
fi

echo "Found PHP versions: ${PHP_VERSIONS[*]}"

for PHP_VERSION in "${PHP_VERSIONS[@]}"; do
    echo "Installing for PHP $PHP_VERSION..."

    # Get PHP paths using the specific version binary
    PHP_BIN="php$PHP_VERSION"
    if ! command -v $PHP_BIN >/dev/null 2>&1; then
        PHP_BIN="php"
    fi

    PHP_EXT_DIR=$($PHP_BIN -i | grep "^extension_dir" | awk '{print $3}')
    PHP_MOD_DIR=$($PHP_BIN -i | grep "Scan this dir for additional .ini files" | awk -F"=> " '{print $2}')

    # Install Aikido PHP extension
    if [ -d "$PHP_EXT_DIR" ]; then
        echo "Installing new Aikido extension in $PHP_EXT_DIR/aikido-%{version}.so..."
        ln -sf /opt/aikido-%{version}/aikido-extension-php-$PHP_VERSION.so $PHP_EXT_DIR/aikido-%{version}.so
    else
        echo "No extension dir for PHP $PHP_VERSION! Skipping..."
        continue
    fi

    # Install Aikido mod
    PHP_DEBIAN_MOD_DIR="/etc/php/$PHP_VERSION/mods-available"
    PHP_DEBIAN_MOD_DIR_CLI="/etc/php/$PHP_VERSION/cli/conf.d"
    PHP_DEBIAN_MOD_DIR_CGI="/etc/php/$PHP_VERSION/cgi/conf.d"
    PHP_DEBIAN_MOD_DIR_FPM="/etc/php/$PHP_VERSION/fpm/conf.d"
    PHP_DEBIAN_MOD_DIR_APACHE2="/etc/php/$PHP_VERSION/apache2/conf.d"

    if [ -d $PHP_DEBIAN_MOD_DIR ]; then
        # Debian-based system
        echo "Installing new Aikido mod in $PHP_DEBIAN_MOD_DIR/aikido-%{version}.ini..."
        ln -sf /opt/aikido-%{version}/aikido.ini $PHP_DEBIAN_MOD_DIR/aikido-%{version}.ini
        if [ -d $PHP_DEBIAN_MOD_DIR_CLI ]; then
            echo "Installing new Aikido mod in $PHP_DEBIAN_MOD_DIR_CLI/zz-aikido-%{version}.ini..."
            ln -sf $PHP_DEBIAN_MOD_DIR/aikido-%{version}.ini $PHP_DEBIAN_MOD_DIR_CLI/zz-aikido-%{version}.ini
        fi
        if [ -d $PHP_DEBIAN_MOD_DIR_CGI ]; then
            echo "Installing new Aikido mod in $PHP_DEBIAN_MOD_DIR_CGI/zz-aikido-%{version}.ini..."
            ln -sf $PHP_DEBIAN_MOD_DIR/aikido-%{version}.ini $PHP_DEBIAN_MOD_DIR_CGI/zz-aikido-%{version}.ini
        fi
        if [ -d $PHP_DEBIAN_MOD_DIR_FPM ]; then
            echo "Installing new Aikido mod in $PHP_DEBIAN_MOD_DIR_FPM/zz-aikido-%{version}.ini..."
            ln -sf $PHP_DEBIAN_MOD_DIR/aikido-%{version}.ini $PHP_DEBIAN_MOD_DIR_FPM/zz-aikido-%{version}.ini
        fi
        if [ -d $PHP_DEBIAN_MOD_DIR_APACHE2 ]; then
            echo "Installing new Aikido mod in $PHP_DEBIAN_MOD_DIR_APACHE2/zz-aikido-%{version}.ini..."
            ln -sf $PHP_DEBIAN_MOD_DIR/aikido-%{version}.ini $PHP_DEBIAN_MOD_DIR_APACHE2/zz-aikido-%{version}.ini
        fi
    else
        # RedHat-based system
        if [ -d "$PHP_MOD_DIR" ]; then
            echo "Installing new Aikido mod in $PHP_MOD_DIR/zz-aikido-%{version}.ini..."
            ln -sf /opt/aikido-%{version}/aikido.ini $PHP_MOD_DIR/zz-aikido-%{version}.ini
        else
            echo "No mod dir for PHP $PHP_VERSION! Skipping..."
            continue
        fi
    fi
done

mkdir -p /run/aikido-%{version}
chmod 777 /run/aikido-%{version}

echo "Installation process for Aikido v%{version} completed."

%preun
#!/bin/bash

echo "Starting the uninstallation process for Aikido v%{version}..."

# Find all PHP versions installed
PHP_VERSIONS=()
if command -v php -v >/dev/null 2>&1; then
    # Add system default PHP version
    PHP_VERSIONS+=($(php -v | grep -oP 'PHP \K\d+\.\d+' | head -n 1))
fi

# Check common PHP installation paths
for php_path in /usr/bin/php* /usr/local/bin/php*; do
    if [[ -x "$php_path" && "$php_path" =~ php([0-9]+\.[0-9]+)$ ]]; then
        version=$("$php_path" -v | grep -oP 'PHP \K\d+\.\d+' | head -n 1)
        if [[ ! " ${PHP_VERSIONS[@]} " =~ " ${version} " ]]; then
            PHP_VERSIONS+=("$version")
        fi
    fi
done

echo "Found PHP versions: ${PHP_VERSIONS[*]}"

for PHP_VERSION in "${PHP_VERSIONS[@]}"; do
    echo "Uninstalling for PHP $PHP_VERSION..."

    # Get PHP paths using the specific version binary
    PHP_BIN="php$PHP_VERSION"
    if ! command -v $PHP_BIN >/dev/null 2>&1; then
        PHP_BIN="php"
    fi

    PHP_EXT_DIR=$($PHP_BIN -i | grep "^extension_dir" | awk '{print $3}')
    PHP_MOD_DIR=$($PHP_BIN -i | grep "Scan this dir for additional .ini files" | awk -F"=> " '{print $2}')

    # Uninstall Aikido mod
    PHP_DEBIAN_MOD_DIR="/etc/php/$PHP_VERSION/mods-available"
    PHP_DEBIAN_MOD_DIR_CLI="/etc/php/$PHP_VERSION/cli/conf.d"
    PHP_DEBIAN_MOD_DIR_CGI="/etc/php/$PHP_VERSION/cgi/conf.d"
    PHP_DEBIAN_MOD_DIR_FPM="/etc/php/$PHP_VERSION/fpm/conf.d"
    PHP_DEBIAN_MOD_DIR_APACHE2="/etc/php/$PHP_VERSION/apache2/conf.d"

    if [ -d $PHP_DEBIAN_MOD_DIR ]; then
        # Debian-based system
        echo "Uninstalling Aikido mod from $PHP_DEBIAN_MOD_DIR/aikido-%{version}.ini..."
        rm -f $PHP_DEBIAN_MOD_DIR/aikido-%{version}.ini
        if [ -d $PHP_DEBIAN_MOD_DIR_CLI ]; then
            echo "Uninstalling Aikido mod from $PHP_DEBIAN_MOD_DIR_CLI/zz-aikido-%{version}.ini..."
            rm -f $PHP_DEBIAN_MOD_DIR_CLI/zz-aikido-%{version}.ini
        fi
        if [ -d $PHP_DEBIAN_MOD_DIR_CGI ]; then
            echo "Uninstalling Aikido mod from $PHP_DEBIAN_MOD_DIR_CGI/zz-aikido-%{version}.ini..."
            rm -f $PHP_DEBIAN_MOD_DIR_CGI/zz-aikido-%{version}.ini
        fi
        if [ -d $PHP_DEBIAN_MOD_DIR_FPM ]; then
            echo "Uninstalling Aikido mod from $PHP_DEBIAN_MOD_DIR_FPM/zz-aikido-%{version}.ini..."
            rm -f $PHP_DEBIAN_MOD_DIR_FPM/zz-aikido-%{version}.ini
        fi
        if [ -d $PHP_DEBIAN_MOD_DIR_APACHE2 ]; then
            echo "Uninstalling Aikido mod from $PHP_DEBIAN_MOD_DIR_APACHE2/zz-aikido-%{version}.ini..."
            rm -f $PHP_DEBIAN_MOD_DIR_APACHE2/zz-aikido-%{version}.ini
        fi
    else
        # RedHat-based system
        if [ -d "$PHP_MOD_DIR" ]; then
            echo "Uninstalling Aikido mod from $PHP_MOD_DIR/zz-aikido-%{version}.ini..."
            rm -f $PHP_MOD_DIR/zz-aikido-%{version}.ini
        fi
    fi

    # Uninstall Aikido PHP extension
    if [ -d "$PHP_EXT_DIR" ]; then
        echo "Uninstalling Aikido extension from $PHP_EXT_DIR/aikido-%{version}.so..."
        rm -f $PHP_EXT_DIR/aikido-%{version}.so
    fi
done

# Remove the Aikido logs folder
rm -rf /var/log/aikido-%{version}

# Remove the Aikido socket folder
SOCKET_FOLDER="/run/aikido-%{version}"

if [ -d "$SOCKET_FOLDER" ]; then
    echo "Removing $SOCKET_FOLDER ..."
    rm -rf "$SOCKET_FOLDER"
    if [ $? -eq 0 ]; then
        echo "Socket folder removed successfully."
    else
        echo "Failed to remove the socket folder."
    fi
else
    echo "Socket $SOCKET_FOLDER does not exist."
fi

echo "Uninstallation process for Aikido v%{version} completed."

%files
/opt/aikido-%{version}

%changelog
* Fri Jun 21 2024 Aikido <hello@aikido.dev> - %{version}-1
- New package version
