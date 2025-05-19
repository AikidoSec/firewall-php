arch=$(uname -m)
VERSION=$(grep '#define PHP_AIKIDO_VERSION' lib/php-extension/include/php_aikido.h | awk -F'"' '{print $2}')

dpkg -i ./aikido-php-firewall-$VERSION-1.$arch.deb