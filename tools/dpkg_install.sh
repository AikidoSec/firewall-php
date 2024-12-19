
VERSION=$(grep '#define PHP_AIKIDO_VERSION' lib/php-extension/include/php_aikido.h | awk -F'"' '{print $2}')

dpkg --purge aikido-php-firewall
alien -i --to-deb --scripts --keep-version /shared/aikido-php-firewall-$VERSION-1.x86_64.rpm
