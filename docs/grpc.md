# gRPC

Similar to RESTful API based on Go kit, the service also supports gRPC transport.
Have a look at [Go kit gRPC example](https://github.com/go-kit/examples/blob/master/addsvc/pkg/addtransport/grpc.go).

The protocol buffer definitions of User and Group services are described in `./rpc/account/account.proto`.
You need to [install](https://grpc.io/docs/quickstart/go.html) protoc compiler and protoc plugins for Go and gRPC to generate code.

```sh
$ brew install protobuf
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

The arguments tell `protoc` to use `account.proto` definition
from within the `./rpc/account/` directory,
generate Go code using Go and gRPC plugins,
and place the result to the `./rpc/` directory.

```sh
$ protoc account.proto --proto_path=./rpc/account/ --go_out=./rpc/ --go-grpc_out=./rpc/

# This also works.
$ protoc ./rpc/account/*.proto --go_out=./rpc/ --go-grpc_out=./rpc/
# The old way no longer works.
$ protoc account.proto -I rpc/account/ --go_out=plugins=grpc:rpc/account/
--go_out: protoc-gen-go: plugins are not supported; use 'protoc --go-grpc_out=...' to generate gRPC
```

We now have newly generated gRPC server and client code in `./rpc/account/account_grpc.pb.go`.

```
$ go doc ./rpc/account/
package account // import "github.com/marselester/ddd-err/rpc/account"

var File_rpc_account_account_proto protoreflect.FileDescriptor
var GroupService_ServiceDesc = grpc.ServiceDesc{ ... }
var UserService_ServiceDesc = grpc.ServiceDesc{ ... }
func RegisterGroupServiceServer(s grpc.ServiceRegistrar, srv GroupServiceServer)
func RegisterUserServiceServer(s grpc.ServiceRegistrar, srv UserServiceServer)
type CreateGroupRequest struct{ ... }
type CreateGroupResponse struct{ ... }
type CreateUserRequest struct{ ... }
type CreateUserResponse struct{ ... }
type Error struct{ ... }
type FindUserByIDRequest struct{ ... }
type FindUserByIDResponse struct{ ... }
type GroupServiceClient interface{ ... }
    func NewGroupServiceClient(cc grpc.ClientConnInterface) GroupServiceClient
type GroupServiceServer interface{ ... }
type UnimplementedGroupServiceServer struct{}
type UnimplementedUserServiceServer struct{}
type UnsafeGroupServiceServer interface{ ... }
type UnsafeUserServiceServer interface{ ... }
type UserServiceClient interface{ ... }
    func NewUserServiceClient(cc grpc.ClientConnInterface) UserServiceClient
type UserServiceServer interface{ ... }
```

## Convention

The convention is inspired by [Twirp best practices](https://twitchtv.github.io/twirp/docs/best_practices.html).
The `.proto` files should follow [Protocol Buffers style guide](https://developers.google.com/protocol-buffers/docs/style).

```
rpc/<domain_name>/<domain_name>.proto
```

For example, the domain package name of this project is `account`.
It defines `UserService` and `GroupService` interfaces.

```
rpc/account/account.proto
```

If this project relied on another one (let's call it "authorization" project),
then its `.proto` file should be copied to corresponding dir.

```
rpc/auth/auth.proto
```

- Use `package <repo_name>.<domain_name>;` for the package name.
- Use `option go_package = "<domain_name>";` for the Go package name.

## grpcurl

[grpcurl](https://github.com/fullstorydev/grpcurl) is like cURL, but for gRPC.

```sh
$ brew install grpcurl
```

Note, your server must support
[reflection](https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md).

```sh
$ go run ./cmd/server/main.go
$ grpcurl -plaintext localhost:8080 list
ddd_err.account.UserService
grpc.reflection.v1alpha.ServerReflection
```

Check if user with "123" ID exists

```sh
$ grpcurl -d '{"id": "123"}' -plaintext localhost:8080 ddd_err.account.UserService/FindUserByID
{
  "error": {
    "message": "Invalid user ID.",
    "code": "invalid_user_id"
  }
}
```

## Buf

[Buf CLI](https://docs.buf.build/tour/introduction) helps to lint proto files, detect breaking changes, and generate code.
There is also Buf Schema Registry in case you don't like manually copying proto files between projects.

```sh
$ brew install bufbuild/buf/buf
```

The `buf.gen.yaml` file controls how the `buf generate` command executes `protoc` plugins.
Here it executes the `protoc-gen-go`, `protoc-gen-go-grpc` plugins and places Go code in the `rpc` directory.

```sh
$ echo 'version: v1
plugins:
  - name: go
    out: rpc
  - name: go-grpc
    out: rpc' > buf.gen.yaml
$ buf generate
```

Verify and lint the proto files.

```sh
$ buf build
$ buf lint
rpc/account/account.proto:2:1:Files with package "ddd_err.account" must be within a directory "ddd_err/account" relative to root but were in directory "rpc/account".
rpc/account/account.proto:2:1:Package name "ddd_err.account" should be suffixed with a correctly formed version, such as "ddd_err.account.v1".
rpc/auth/auth.proto:2:1:Files with package "ddd_auth.auth" must be within a directory "ddd_auth/auth" relative to root but were in directory "rpc/auth".
rpc/auth/auth.proto:2:1:Package name "ddd_auth.auth" should be suffixed with a correctly formed version, such as "ddd_auth.auth.v1".
```

You can set lint exceptions in `buf.yaml`.
The placement of the `buf.yaml` is analogous to a `protoc ... -I rpc/account/`.

```sh
$ cd ./rpc/account/
$ buf mod init # It creates ./rpc/account/buf.yaml.
$ cat buf.yaml
version: v1
breaking:
  use:
    - FILE
lint:
  use:
    - DEFAULT
```
