rm -rf build
mkdir build
cd src
protoc --go_out=agent --go-grpc_out=agent ipc.proto
protoc --go_out=lib --go-grpc_out=lib ipc.proto
cd agent
go get google.golang.org/grpc
go build -o ../../build/aikido
cd ../extension
phpize
cd ../lib
go get google.golang.org/grpc
go build -buildmode=c-archive -o ../../build/libaikido_go.a
cd ../../build
CXX=g++ CXXFLAGS="-fPIC -std=c++20 -g -O0 -I../include" LDFLAGS="-L./ -laikido_go -lstdc++" ../src/extension/configure
make
cd ..

sudo cp -f ./build/modules/aikido.so /opt/aikido/aikido-1.4.0-extension-php-8.0.so 