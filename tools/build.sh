export PATH="$PATH:$HOME/go/bin:$HOME/.local/bin"

PHP_VERSION=$(php -v | grep -oP 'PHP \K\d+\.\d+' | head -n 1)
AIKIDO_EXTENSION=aikido-extension-php-$PHP_VERSION.so 
AIKIDO_EXTENSION_DEBUG=aikido-extension-php-$PHP_VERSION.so.debug

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
go test ./...
go build -ldflags "-s -w" -o ../../build/aikido-agent
cd ../request-processor
go get google.golang.org/grpc
go get github.com/stretchr/testify/assert
go test ./...
go build -ldflags "-s -w" -buildmode=c-shared  -o ../../build/aikido-request-processor.so
cd ../../build
CXX=g++ CXXFLAGS="-fPIC -g -O2 -I../lib/php-extension/include" LDFLAGS="-lstdc++" ../lib/php-extension/configure
sed -i "s/available_tags=''/available_tags='CXX'/" libtool
if ! grep -q "BEGIN LIBTOOL TAG CONFIG: CXX" libtool; then
  sed -i '/^# ### BEGIN LIBTOOL TAG CONFIG: disable-shared$/i\
# ### BEGIN LIBTOOL TAG CONFIG: CXX\
LTCXX="g++"\
CXXFLAGS="-fPIC -g -O2"\
compiler_CXX="g++"\
# ### END LIBTOOL TAG CONFIG: CXX\
' libtool
fi
make
cd ./modules/
mv aikido.so $AIKIDO_EXTENSION
objcopy --only-keep-debug $AIKIDO_EXTENSION $AIKIDO_EXTENSION_DEBUG
strip --strip-debug $AIKIDO_EXTENSION
objcopy --add-gnu-debuglink=$AIKIDO_EXTENSION_DEBUG $AIKIDO_EXTENSION

cd ../..
