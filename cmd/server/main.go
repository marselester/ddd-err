// Server exposes REST-style and gRPC API to manage user accounts.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/log"
	"github.com/oklog/run"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	account "github.com/marselester/ddd-err"
	"github.com/marselester/ddd-err/api"
	"github.com/marselester/ddd-err/mock"
	pb "github.com/marselester/ddd-err/rpc/account"
)

func main() {
	apiAddr := flag.String("http", ":8000", "HTTP API address")
	apiQPS := flag.Int("qps", 2, "API requests limit per second")
	grpcAddr := flag.String("grpc", ":8080", "gRPC API address")
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	// db helps to emulate storage errors.
	db := &mock.UserStorage{
		FindUserByIDFn: func(ctx context.Context, dbtx *sql.Tx, id string) (*account.User, error) {
			return nil, account.Error{
				Code:    account.ENotFound,
				Message: "User not found.",
				Inner:   sql.ErrNoRows,
			}
		},
		UsernameInUseFn: func(ctx context.Context, username string) bool {
			return username == "bob"
		},
		CreateUserFn: func(ctx context.Context, user *account.User) error {
			return fmt.Errorf(
				"UserStorage.CreateUser: %w",
				fmt.Errorf(
					"insertUser: %w",
					fmt.Errorf("db connection failed"),
				),
			)
		},
	}

	var s account.UserService
	{
		s = api.NewService(
			db,
			api.WithLogger(logger),
		)
		s = api.NewLoggingMiddleware(logger, s)
	}

	// REST-style API server for creating new users.
	apiserver := http.Server{
		Addr: *apiAddr,
		Handler: api.NewHTTPHandler(
			s,
			log.With(logger, "component", "HTTP"),
			*apiQPS,
		),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// gRPC API server for creating new users.
	grpcListener, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		logger.Log("msg", "could not listen to gRPC port", "err", err)
		os.Exit(1)
	}
	grpcserver := grpc.NewServer()
	pb.RegisterUserServiceServer(
		grpcserver,
		api.NewGRPCUserServer(s, logger, *apiQPS),
	)
	// gRPC reflection provides information about publicly-accessible gRPC services on a server,
	// and assists clients at runtime to construct RPC requests and responses
	// without precompiled service information. It is used by grpcurl CLI.
	reflection.Register(grpcserver)

	ctx, cancel := context.WithCancel(context.Background())
	var g run.Group
	{
		g.Add(func() error {
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			return fmt.Errorf("signal received: %v", <-sig)
		}, func(err error) {
			logger.Log("msg", "program was interrupted", "err", err)
			cancel()
		})
	}
	{
		g.Add(func() error {
			logger.Log("msg", "API server is starting", "addr", *apiAddr)
			return apiserver.ListenAndServe()
		}, func(err error) {
			logger.Log("msg", "API server was interrupted", "err", err)
			err = apiserver.Shutdown(ctx)
			logger.Log("msg", "API server shut down", "err", err)
		})
	}
	{
		g.Add(func() error {
			logger.Log("msg", "gRPC server is starting", "addr", *grpcAddr)
			return grpcserver.Serve(grpcListener)
		}, func(err error) {
			logger.Log("msg", "gRPC server was interrupted", "err", err)
			grpcserver.GracefulStop()
			logger.Log("msg", "gRPC server shut down")
		})
	}
	err = g.Run()
	logger.Log("msg", "actors stopped", "err", err)
}
