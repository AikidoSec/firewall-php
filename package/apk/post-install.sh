#!/bin/sh
set -e

VERSION="__VERSION__"

echo "Starting the installation process for Aikido PHP Firewall v${VERSION}..."

pids=$(ps aux 2>/dev/null | grep aikido-agent | grep -v grep | awk '{print $2}') || true
if [ -n "$pids" ]; then
    echo "Stopping Aikido Agent processes: $pids"
    echo "$pids" | xargs kill -9 2>/dev/null || true
    echo "Aikido Agent(s) stopped."
fi

mkdir -p /var/log/aikido-${VERSION}
chmod 777 /var/log/aikido-${VERSION}

PHP_VERSIONS=""
if command -v php >/dev/null 2>&1; then
    ver=$(php -r 'echo PHP_MAJOR_VERSION . "." . PHP_MINOR_VERSION;')
    PHP_VERSIONS="$ver"
fi

for php_path in /usr/bin/php[0-9]* /usr/local/bin/php[0-9]*; do
    if [ -x "$php_path" ]; then
        ver=$("$php_path" -r 'echo PHP_MAJOR_VERSION . "." . PHP_MINOR_VERSION;' 2>/dev/null) || continue
        case " $PHP_VERSIONS " in
            *" $ver "*) ;;
            *) PHP_VERSIONS="$PHP_VERSIONS $ver" ;;
        esac
    fi
done

PHP_VERSIONS=$(echo "$PHP_VERSIONS" | xargs)

if [ -n "$PHP_VERSIONS" ]; then
    echo "Found PHP versions: $PHP_VERSIONS"
fi

FRANKENPHP_PHP_VERSION=""
if command -v frankenphp >/dev/null 2>&1; then
    FRANKENPHP_PHP_VERSION=$(frankenphp -v 2>/dev/null | sed -n 's/.*PHP \([0-9]*\.[0-9]*\).*/\1/p' | head -1)
    if [ -n "$FRANKENPHP_PHP_VERSION" ]; then
        echo "Found FrankenPHP with embedded PHP $FRANKENPHP_PHP_VERSION"
    fi
fi

for PHP_VERSION in $PHP_VERSIONS; do
    echo "Installing for PHP $PHP_VERSION..."

    PHP_BIN="php${PHP_VERSION}"
    if ! command -v "$PHP_BIN" >/dev/null 2>&1; then
        PHP_BIN="php"
    fi

    PHP_EXT_DIR=$("$PHP_BIN" -i | grep "^extension_dir" | awk '{print $3}')
    PHP_MOD_DIR=$("$PHP_BIN" -i | grep "Scan this dir for additional .ini files" | awk -F"=> " '{print $2}')

    PHP_THREAD_SAFETY=$("$PHP_BIN" -i | grep "Thread Safety" | awk -F"=> " '{print $2}' | tr -d ' ')
    if [ "$PHP_THREAD_SAFETY" = "enabled" ]; then
        EXT_SUFFIX="-zts"
        echo "PHP $PHP_VERSION is ZTS (Thread Safe)"
    else
        EXT_SUFFIX="-nts"
        echo "PHP $PHP_VERSION is NTS (Non-Thread Safe)"
    fi

    if [ -d "$PHP_EXT_DIR" ]; then
        EXT_FILE="aikido-extension-php-${PHP_VERSION}${EXT_SUFFIX}.so"
        if [ -f "/opt/aikido-${VERSION}/${EXT_FILE}" ]; then
            echo "Installing new Aikido extension in ${PHP_EXT_DIR}/aikido-${VERSION}.so..."
            ln -sf "/opt/aikido-${VERSION}/${EXT_FILE}" "${PHP_EXT_DIR}/aikido-${VERSION}.so"
        else
            echo "Warning: Extension file /opt/aikido-${VERSION}/${EXT_FILE} not found! Skipping..."
            continue
        fi
    else
        echo "No extension dir for PHP ${PHP_VERSION}! Skipping..."
        continue
    fi

    PHP_DEBIAN_MOD_DIR="/etc/php/${PHP_VERSION}/mods-available"
    PHP_DEBIAN_MOD_DIR_CLI="/etc/php/${PHP_VERSION}/cli/conf.d"
    PHP_DEBIAN_MOD_DIR_CGI="/etc/php/${PHP_VERSION}/cgi/conf.d"
    PHP_DEBIAN_MOD_DIR_FPM="/etc/php/${PHP_VERSION}/fpm/conf.d"
    PHP_DEBIAN_MOD_DIR_APACHE2="/etc/php/${PHP_VERSION}/apache2/conf.d"

    if [ -d "$PHP_DEBIAN_MOD_DIR" ]; then
        echo "Installing new Aikido mod in ${PHP_DEBIAN_MOD_DIR}/aikido-${VERSION}.ini..."
        ln -sf "/opt/aikido-${VERSION}/aikido.ini" "${PHP_DEBIAN_MOD_DIR}/aikido-${VERSION}.ini"
        for subdir in "$PHP_DEBIAN_MOD_DIR_CLI" "$PHP_DEBIAN_MOD_DIR_CGI" "$PHP_DEBIAN_MOD_DIR_FPM" "$PHP_DEBIAN_MOD_DIR_APACHE2"; do
            if [ -d "$subdir" ]; then
                echo "Installing new Aikido mod in ${subdir}/zz-aikido-${VERSION}.ini..."
                ln -sf "${PHP_DEBIAN_MOD_DIR}/aikido-${VERSION}.ini" "${subdir}/zz-aikido-${VERSION}.ini"
            fi
        done
    elif [ -d "$PHP_MOD_DIR" ]; then
        echo "Installing new Aikido mod in ${PHP_MOD_DIR}/zz-aikido-${VERSION}.ini..."
        ln -sf "/opt/aikido-${VERSION}/aikido.ini" "${PHP_MOD_DIR}/zz-aikido-${VERSION}.ini"
    else
        echo "No mod dir for PHP ${PHP_VERSION}! Skipping..."
        continue
    fi
done

if [ -n "$FRANKENPHP_PHP_VERSION" ]; then
    echo "Installing for FrankenPHP with PHP ${FRANKENPHP_PHP_VERSION}... ZTS (Thread Safe)"

    FRANKENPHP_EXT_DIR="/usr/lib/frankenphp/modules"
    FRANKENPHP_INI_DIR="/etc/frankenphp/php.d"

    mkdir -p "$FRANKENPHP_EXT_DIR" "$FRANKENPHP_INI_DIR"
    ln -sf "/opt/aikido-${VERSION}/aikido-extension-php-${FRANKENPHP_PHP_VERSION}-zts.so" "${FRANKENPHP_EXT_DIR}/aikido-${VERSION}.so"
    ln -sf "/opt/aikido-${VERSION}/aikido.ini" "${FRANKENPHP_INI_DIR}/zz-aikido-${VERSION}.ini"
fi

if [ -z "$PHP_VERSIONS" ] && [ -z "$FRANKENPHP_PHP_VERSION" ]; then
    echo "No PHP or FrankenPHP found! Exiting!"
    exit 1
fi

mkdir -p /run/aikido-${VERSION}
chmod 777 /run/aikido-${VERSION}

echo "Installation process for Aikido v${VERSION} completed."
