VERSION=$(grep '#define PHP_AIKIDO_VERSION' lib/php-extension/include/php_aikido.h | awk -F'"' '{print $2}')

rpm -Uvh --oldpackage ~/rpmbuild/RPMS/x86_64/aikido-php-firewall-$VERSION-1.x86_64.rpm
cp ~/rpmbuild/RPMS/x86_64/aikido-php-firewall-$VERSION-1.x86_64.rpm /shared
