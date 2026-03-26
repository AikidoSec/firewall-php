#!/bin/sh
set -e

VERSION="__VERSION__"

echo "Starting the uninstallation process for Aikido v${VERSION}..."

pids=$(ps aux 2>/dev/null | grep aikido-agent | grep -v grep | awk '{print $2}') || true
if [ -n "$pids" ]; then
    echo "Stopping Aikido Agent processes: $pids"
    echo "$pids" | xargs kill -15 2>/dev/null || true
    echo "Aikido Agent(s) stopped."
fi

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

echo "Found PHP versions: $PHP_VERSIONS"

FRANKENPHP_EXT_DIR="/usr/lib/frankenphp/modules"
FRANKENPHP_INI_DIR="/etc/frankenphp/php.d"

for PHP_VERSION in $PHP_VERSIONS; do
    echo "Uninstalling for PHP ${PHP_VERSION}..."

    PHP_BIN="php${PHP_VERSION}"
    if ! command -v "$PHP_BIN" >/dev/null 2>&1; then
        PHP_BIN="php"
    fi

    PHP_EXT_DIR=$("$PHP_BIN" -i | grep "^extension_dir" | awk '{print $3}')
    PHP_MOD_DIR=$("$PHP_BIN" -i | grep "Scan this dir for additional .ini files" | awk -F"=> " '{print $2}')

    PHP_DEBIAN_MOD_DIR="/etc/php/${PHP_VERSION}/mods-available"
    PHP_DEBIAN_MOD_DIR_CLI="/etc/php/${PHP_VERSION}/cli/conf.d"
    PHP_DEBIAN_MOD_DIR_CGI="/etc/php/${PHP_VERSION}/cgi/conf.d"
    PHP_DEBIAN_MOD_DIR_FPM="/etc/php/${PHP_VERSION}/fpm/conf.d"
    PHP_DEBIAN_MOD_DIR_APACHE2="/etc/php/${PHP_VERSION}/apache2/conf.d"

    if [ -d "$PHP_DEBIAN_MOD_DIR" ]; then
        echo "Uninstalling Aikido mod from ${PHP_DEBIAN_MOD_DIR}/aikido-${VERSION}.ini..."
        rm -f "${PHP_DEBIAN_MOD_DIR}/aikido-${VERSION}.ini"
        for subdir in "$PHP_DEBIAN_MOD_DIR_CLI" "$PHP_DEBIAN_MOD_DIR_CGI" "$PHP_DEBIAN_MOD_DIR_FPM" "$PHP_DEBIAN_MOD_DIR_APACHE2"; do
            if [ -d "$subdir" ]; then
                rm -f "${subdir}/zz-aikido-${VERSION}.ini"
            fi
        done
    elif [ -d "$PHP_MOD_DIR" ]; then
        echo "Uninstalling Aikido mod from ${PHP_MOD_DIR}/zz-aikido-${VERSION}.ini..."
        rm -f "${PHP_MOD_DIR}/zz-aikido-${VERSION}.ini"
    fi

    if [ -d "$PHP_EXT_DIR" ]; then
        echo "Uninstalling Aikido extension from ${PHP_EXT_DIR}/aikido-${VERSION}.so..."
        rm -f "${PHP_EXT_DIR}/aikido-${VERSION}.so"
    fi
done

if [ -d "$FRANKENPHP_EXT_DIR" ] || [ -d "$FRANKENPHP_INI_DIR" ]; then
    echo "Uninstalling for FrankenPHP..."
    rm -f "${FRANKENPHP_EXT_DIR}/aikido-${VERSION}.so" 2>/dev/null || true
    rm -f "${FRANKENPHP_INI_DIR}/zz-aikido-${VERSION}.ini" 2>/dev/null || true
fi

rm -rf "/var/log/aikido-${VERSION}"

SOCKET_FOLDER="/run/aikido-${VERSION}"
if [ -d "$SOCKET_FOLDER" ]; then
    echo "Removing $SOCKET_FOLDER ..."
    rm -rf "$SOCKET_FOLDER"
fi

echo "Uninstallation process for Aikido v${VERSION} completed."
