export PATH="/opt/homebrew/opt/llvm/bin:$PATH"

PHP_VERSION=$(php -v | ggrep -oP 'PHP \K\d+\.\d+' | head -n 1)
AIKIDO_EXTENSION=aikido-extension-php-$PHP_VERSION.dylib 
AIKIDO_EXTENSION_DEBUG=aikido-extension-php-$PHP_VERSION.dylib.debug

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
go build -ldflags "-s -w" -buildmode=c-shared  -o ../../build/aikido-agent.dylib
cd ../request-processor
go get google.golang.org/grpc
go get github.com/stretchr/testify/assert
go get github.com/seancfoley/ipaddress-go/ipaddr
go test ./...
go build -ldflags "-s -w" -buildmode=c-shared  -o ../../build/aikido-request-processor.dylib
cd ../../build
CXX=clang CXXFLAGS="-fPIC -g -O2 -I../lib/php-extension/include -std=c++17" LDFLAGS="-lstdc++" ../lib/php-extension/configure
make
cd ./modules/
mv aikido.so $AIKIDO_EXTENSION

cp $AIKIDO_EXTENSION /opt/homebrew/lib/php/pecl/20210902/$AIKIDO_EXTENSION
# llvm-objcopy --only-keep-debug $AIKIDO_EXTENSION $AIKIDO_EXTENSION_DEBUG
# llvm-objcopy --strip-debug $AIKIDO_EXTENSION
# llvm-objcopy --add-gnu-debuglink=$AIKIDO_EXTENSION_DEBUG $AIKIDO_EXTENSION

cd ../..
