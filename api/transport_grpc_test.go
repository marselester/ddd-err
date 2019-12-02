package api_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/go-kit/kit/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	account "github.com/marselester/ddd-err"
	"github.com/marselester/ddd-err/api"
	"github.com/marselester/ddd-err/apiclient"
	"github.com/marselester/ddd-err/mock"
	pb "github.com/marselester/ddd-err/rpc/account"
)

func TestGRPCUserService_ratelimit(t *testing.T) {
	svc := api.NewService(nil)
	usrSrv := api.NewGRPCUserServer(svc, log.NewNopLogger(), 1)

	grpcListener := bufconn.Listen(1024)
	grpcserver := grpc.NewServer()
	pb.RegisterUserServer(grpcserver, usrSrv)
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
	svc = apiclient.NewGRPCUserClient(conn)

	userID := "123"
	svc.FindUserByID(context.Background(), userID)

	user, err := svc.FindUserByID(context.Background(), userID)
	want := account.Error{
		Code:    "rate_limit",
		Message: "API rate limit exceeded.",
	}
	if !errors.Is(err, want) {
		t.Errorf("FindUserByID(%q) = %q want %q", userID, err, want)
	}
	if user != nil {
		t.Errorf("FindUserByID(%q) = %+v want nil", userID, user)
	}
}

func TestGRPCUserService_FindUserByID_invalid_user_id(t *testing.T) {
	svc := api.NewService(nil)
	usrSrv := api.NewGRPCUserServer(svc, log.NewNopLogger(), 100)

	grpcListener := bufconn.Listen(1024)
	grpcserver := grpc.NewServer()
	pb.RegisterUserServer(grpcserver, usrSrv)
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
	svc = apiclient.NewGRPCUserClient(conn)

	userID := "123"
	user, err := svc.FindUserByID(context.Background(), userID)
	want := account.Error{
		Code:    "invalid_user_id",
		Message: "Invalid user ID.",
	}
	if !errors.Is(err, want) {
		t.Errorf("FindUserByID(%q) = %q want %q", userID, err, want)
	}
	if user != nil {
		t.Errorf("FindUserByID(%q) = %+v want nil", userID, user)
	}
}

func TestGRPCUserService_FindUserByID_notfound(t *testing.T) {
	svc := api.NewService(&mock.UserStorage{
		FindUserByIDFn: func(ctx context.Context, dbtx *sql.Tx, id string) (*account.User, error) {
			return nil, account.Error{
				Code:    account.ENotFound,
				Message: "User not found.",
				Inner:   sql.ErrNoRows,
			}
		},
	})
	usrSrv := api.NewGRPCUserServer(svc, log.NewNopLogger(), 100)

	grpcListener := bufconn.Listen(1024)
	grpcserver := grpc.NewServer()
	pb.RegisterUserServer(grpcserver, usrSrv)
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
	svc = apiclient.NewGRPCUserClient(conn)

	userID := "87553f14-4c0f-4bd8-8be1-1b6ff5bd8eef"
	user, err := svc.FindUserByID(context.Background(), userID)
	want := account.Error{
		Code:    "not_found",
		Message: "User not found.",
	}
	if !errors.Is(err, want) {
		t.Errorf("FindUserByID(%q) = %q want %q", userID, err, want)
	}
	if user != nil {
		t.Errorf("FindUserByID(%q) = %+v want nil", userID, user)
	}
}

func TestGRPCUserService_CreateUser_validation(t *testing.T) {
	tt := []struct {
		user account.User
		want account.Error
	}{
		{
			account.User{},
			account.Error{Code: "invalid_username", Message: "Username is invalid."},
		},
		{
			account.User{Username: " "},
			account.Error{Code: "invalid_username", Message: "Username is invalid."},
		},
		{
			account.User{Username: ">_<"},
			account.Error{Code: "invalid_username", Message: "Username is invalid."},
		},
		{
			account.User{Username: "bob123"},
			account.Error{Code: "conflict", Message: "Username is already in use. Please choose a different username."},
		},
	}

	svc := api.NewService(&mock.UserStorage{})
	usrSrv := api.NewGRPCUserServer(svc, log.NewNopLogger(), 100)

	grpcListener := bufconn.Listen(1024)
	grpcserver := grpc.NewServer()
	pb.RegisterUserServer(grpcserver, usrSrv)
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
	svc = apiclient.NewGRPCUserClient(conn)

	for _, tc := range tt {
		err = svc.CreateUser(context.Background(), &tc.user)
		if !errors.Is(err, tc.want) {
			t.Errorf("CreateUser(%+v) = %q want %q", tc.user, err, tc.want)
		}
	}
}

func TestGRPCUserService_CreateUser_dberror(t *testing.T) {
	svc := api.NewService(&mock.UserStorage{
		UsernameInUseFn: func(ctx context.Context, username string) bool {
			return false
		},
		CreateUserFn: func(ctx context.Context, user *account.User) error {
			return fmt.Errorf("UserStorage.CreateUser: %w", errors.New("db connection failed"))
		},
	})
	usrSrv := api.NewGRPCUserServer(svc, log.NewNopLogger(), 100)

	grpcListener := bufconn.Listen(1024)
	grpcserver := grpc.NewServer()
	pb.RegisterUserServer(grpcserver, usrSrv)
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
	svc = apiclient.NewGRPCUserClient(conn)

	user := account.User{
		Username: "bob123",
	}
	err = svc.CreateUser(context.Background(), &user)
	want := account.Error{
		Code:    "internal",
		Message: "An internal error has occurred.",
	}
	if !errors.Is(err, want) {
		t.Errorf("CreateUser(%+v) = %q want %q", user, err, want)
	}
}
