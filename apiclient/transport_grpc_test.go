package apiclient_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	account "github.com/marselester/ddd-err"
	"github.com/marselester/ddd-err/api"
	"github.com/marselester/ddd-err/apiclient"
	"github.com/marselester/ddd-err/mock"
	pb "github.com/marselester/ddd-err/rpc/account"
)

func TestGRPCUserService_circuitbreaker(t *testing.T) {
	tt := []struct {
		name string
		err  account.Error
		want error
	}{
		{
			name: "validation error",
			err:  account.Error{Code: "invalid_username", Message: "Username is invalid."},
			want: account.Error{Code: "invalid_username", Message: "Username is invalid."},
		},
		{
			name: "too many requests",
			err:  account.Error{Code: "rate_limit", Message: "API rate limit exceeded."},
			want: gobreaker.ErrOpenState,
		},
		{
			name: "server error",
			err:  account.Error{Code: "internal", Message: "An internal error has occurred."},
			want: gobreaker.ErrOpenState,
		},
	}

	svc := mock.UserService{}
	usrSrv := api.NewGRPCUserServer(&svc, log.NewNopLogger(), 100)

	grpcListener := bufconn.Listen(1024)
	grpcserver := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcserver, usrSrv)
	go func() {
		if err := grpcserver.Serve(grpcListener); err != nil {
			t.Errorf("grpc serve failed: %v", err)
		}
	}()
	defer grpcserver.Stop()

	conn, err := grpc.Dial("", grpc.WithInsecure(), grpc.WithContextDialer(
		func(context.Context, string) (net.Conn, error) {
			return grpcListener.Dial()
		}),
	)
	if err != nil {
		t.Fatalf("grpc dial failed: %v", err)
	}
	defer conn.Close()
	c := apiclient.NewGRPCUserClient(conn)

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc.CreateUserFn = func(ctx context.Context, user *account.User) error {
				return tc.err
			}

			for i := 0; i < 7; i++ {
				err = c.CreateUser(context.Background(), &account.User{})
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("CreateUser() = %q want %q", err, tc.want)
			}
		})
	}
}
