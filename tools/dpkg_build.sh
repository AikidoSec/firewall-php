arch=$(uname -m)
VERSION=$(grep '#define PHP_AIKIDO_VERSION' lib/php-extension/include/php_aikido.h | awk -F'"' '{print $2}')

alien --to-deb --scripts --keep-version /shared/aikido-php-firewall-$VERSION-1.$arch.rpm
mv aikido-php-firewall_$VERSION-1_amd64.deb temp-aikido-php-firewall-$VERSION-1.$arch.deb

mkdir deb-temp
dpkg-deb -R temp-aikido-php-firewall-$VERSION-1.$arch.deb deb-temp/
dpkg-deb -Zgzip -b deb-temp aikido-php-firewall-$VERSION-1.$arch.deb
rm -rf deb-temp

cp aikido-php-firewall-$VERSION-1.$arch.deb .devcontainer/shared/
