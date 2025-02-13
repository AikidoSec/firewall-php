rm -rf ~/rpmbuild
rpmdev-setuptree

arch=$(uname -m)

PHP_VERSION=$(php -v | grep -oP 'PHP \K\d+\.\d+' | head -n 1)
VERSION=$(grep '#define PHP_AIKIDO_VERSION' lib/php-extension/include/php_aikido.h | awk -F'"' '{print $2}')
AIKIDO_INTERNALS_REPO=https://api.github.com/repos/AikidoSec/zen-internals
AIKIDO_INTERNALS_LIB=libzen_internals_$arch-unknown-linux-gnu.so

mkdir -p ~/rpmbuild/SOURCES/aikido-php-firewall-$VERSION

cp -rf package/rpm/opt ~/rpmbuild/SOURCES/aikido-php-firewall-$VERSION/

cp -f package/rpm/aikido.spec ~/rpmbuild/SPECS/

cp build/aikido-agent.so ~/rpmbuild/SOURCES/aikido-php-firewall-$VERSION/opt/aikido/aikido-agent.so
cp build/aikido-request-processor.so ~/rpmbuild/SOURCES/aikido-php-firewall-$VERSION/opt/aikido/aikido-request-processor.so
cp build/modules/aikido-extension-php-$PHP_VERSION.so ~/rpmbuild/SOURCES/aikido-php-firewall-$VERSION/opt/aikido/aikido-extension-php-$PHP_VERSION.so

curl -L -o $AIKIDO_INTERNALS_LIB $(curl -s $AIKIDO_INTERNALS_REPO/releases/latest | jq -r ".assets[] | select(.name == \"$AIKIDO_INTERNALS_LIB\") | .browser_download_url")
mv $AIKIDO_INTERNALS_LIB ~/rpmbuild/SOURCES/aikido-php-firewall-$VERSION/opt/aikido/$AIKIDO_INTERNALS_LIB

mv ~/rpmbuild/SOURCES/aikido-php-firewall-$VERSION/opt/aikido ~/rpmbuild/SOURCES/aikido-php-firewall-$VERSION/opt/aikido-$VERSION

sed -i "s/aikido.so/aikido-$VERSION.so/" ~/rpmbuild/SOURCES/aikido-php-firewall-$VERSION/opt/aikido-$VERSION/aikido.ini

cd ~/rpmbuild/SOURCES
tar czvf ~/rpmbuild/SOURCES/aikido-php-firewall-$VERSION.tar.gz *
rm -rf ~/rpmbuild/SOURCES/aikido-php-firewall-$VERSION

rpmbuild --define "debug_package %{nil}" -ba ~/rpmbuild/SPECS/aikido.spec
