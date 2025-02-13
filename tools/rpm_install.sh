arch=$(uname -m)
VERSION=$(grep '#define PHP_AIKIDO_VERSION' lib/php-extension/include/php_aikido.h | awk -F'"' '{print $2}')

rpm -Uvh --oldpackage ~/rpmbuild/RPMS/$arch/aikido-php-firewall-$VERSION-1.$arch.rpm
cp ~/rpmbuild/RPMS/$arch/aikido-php-firewall-$VERSION-1.$arch.rpm /shared
