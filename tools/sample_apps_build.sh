set -e

arch=$(uname -m)

export PATH="$PATH:$HOME/go/bin:$HOME/.local/bin"

PHP_VERSION=$(php -v | grep -oP 'PHP \K\d+\.\d+' | head -n 1)
AIKIDO_VERSION=$(grep '#define PHP_AIKIDO_VERSION' lib/php-extension/include/php_aikido.h | awk -F'"' '{print $2}')
AIKIDO_EXTENSION=aikido-extension-php-$AIKIDO_VERSION.so 
AIKIDO_EXTENSION_DEBUG=aikido-extension-php-$AIKIDO_VERSION.so.debug
AIKIDO_INTERNALS_REPO=https://api.github.com/repos/AikidoSec/zen-internals
AIKIDO_INTERNALS_LIB=libzen_internals_$arch-unknown-linux-gnu.so


rm -rf build
mkdir build
cd lib
cd php-extension
phpize
cd ..
protoc --go_out=agent --go-grpc_out=agent ipc.proto
protoc --go_out=request-processor --go-grpc_out=request-processor ipc.proto
cd agent
go get google.golang.org/grpc
go mod tidy
go build -ldflags "-s -w" -o ../../build/aikido-agent
cd ../request-processor
go get google.golang.org/grpc
go get github.com/stretchr/testify/assert
go mod tidy
go build -ldflags "-s -w" -buildmode=c-shared  -o ../../build/aikido-request-processor.so
cd ../../build
CXX=g++ CXXFLAGS="-fPIC -g -O2 -I../lib/php-extension/include" LDFLAGS="-lstdc++" ../lib/php-extension/configure
make
cd ./modules/
mv aikido.so $AIKIDO_EXTENSION
objcopy --only-keep-debug $AIKIDO_EXTENSION $AIKIDO_EXTENSION_DEBUG
strip --strip-debug $AIKIDO_EXTENSION
objcopy --add-gnu-debuglink=$AIKIDO_EXTENSION_DEBUG $AIKIDO_EXTENSION

cd ../..

AIKIDO_INSTALL_PATH="/opt/aikido-$AIKIDO_VERSION"
mkdir -p $AIKIDO_INSTALL_PATH
cp build/aikido-agent $AIKIDO_INSTALL_PATH/
cp build/aikido-request-processor.so $AIKIDO_INSTALL_PATH/
cp build/modules/$AIKIDO_EXTENSION $(php -r 'echo ini_get("extension_dir");')/aikido.so

curl -L -o $AIKIDO_INTERNALS_LIB $(curl -s $AIKIDO_INTERNALS_REPO/releases/latest | jq -r ".assets[] | select(.name == \"$AIKIDO_INTERNALS_LIB\") | .browser_download_url")
mv $AIKIDO_INTERNALS_LIB $AIKIDO_INSTALL_PATH/$AIKIDO_INTERNALS_LIB

chmod 755 $AIKIDO_INSTALL_PATH
chmod 644 $AIKIDO_INSTALL_PATH/aikido-agent
chmod 644 $AIKIDO_INSTALL_PATH/aikido-request-processor.so
chmod 644 $(php -r 'echo ini_get("extension_dir");')/aikido.so

echo "extension=aikido.so" | tee /etc/php/$PHP_VERSION/mods-available/aikido.ini
phpenmod aikido

ldconfig

mkdir -p /var/log/aikido-$AIKIDO_VERSION
chmod 644 /var/log/aikido-$AIKIDO_VERSION

php -m | grep aikido || echo "Aikido extension not loaded!"