# Protocol Buffers definitions

`empty.proto` is copied unchanged from [Protocol Buffers v28.3](https://github.com/protocolbuffers/protobuf/blob/5fda5abda3dee5f7a102c85860594bff8d8610bd/src/google/protobuf/empty.proto), including its license notice.

Our generation commands run from `lib`, so `protoc` finds this copy before the system definitions. It keeps the Go import pointing to `google.golang.org/protobuf/types/known/emptypb`; Ubuntu 20.04's system copy points to the legacy `github.com/golang/protobuf` module.
