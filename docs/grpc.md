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

func RegisterGroupServer(s *grpc.Server, srv GroupServer)
func RegisterUserServer(s *grpc.Server, srv UserServer)
type CreateGroupReq struct{ ... }
type CreateGroupResp struct{ ... }
type CreateUserReq struct{ ... }
type CreateUserResp struct{ ... }
type Error struct{ ... }
type FindUserByIDReq struct{ ... }
type FindUserByIDResp struct{ ... }
type GroupClient interface{ ... }
    func NewGroupClient(cc *grpc.ClientConn) GroupClient
type GroupServer interface{ ... }
type UnimplementedGroupServer struct{}
type UnimplementedUserServer struct{}
type UserClient interface{ ... }
    func NewUserClient(cc *grpc.ClientConn) UserClient
type UserServer interface{ ... }
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
Install it from source

```sh
$ go get github.com/fullstorydev/grpcurl
$ go install github.com/fullstorydev/grpcurl/cmd/grpcurl
```

and you're ready to send requests to the gRPC server. Note, your server must support
[reflection](https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md).

```sh
$ go run ./cmd/server/main.go
$ grpcurl -plaintext localhost:8080 list
ddd_err.account.User
grpc.reflection.v1alpha.ServerReflection
```

Check if user with "123" ID exists

```sh
$ grpcurl -d '{"id": "123"}' -plaintext localhost:8080 ddd_err.account.User/FindUserByID
{
  "error": {
    "message": "Invalid user ID.",
    "code": "invalid_user_id"
  }
}
```
