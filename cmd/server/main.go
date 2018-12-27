// Server exposes REST-style API to manage user accounts.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/oklog/run"

	account "github.com/marselester/ddd-err"
	"github.com/marselester/ddd-err/api"
	"github.com/marselester/ddd-err/mock"
)

func main() {
	apiAddr := flag.String("http", ":8000", "HTTP API address")
	apiQPS := flag.Int("qps", 2, "API requests limit per second")
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	// db helps to emulate storage errors.
	db := &mock.UserStorage{
		UsernameInUseFn: func(ctx context.Context, username string) bool {
			return username == "bob"
		},
		CreateUserFn: func(ctx context.Context, user *account.User) error {
			return &account.Error{
				Op: "UserStorage.CreateUser",
				Err: &account.Error{
					Op:  "insertUser",
					Err: fmt.Errorf("db connection failed"),
				},
			}
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
	err := g.Run()
	logger.Log("msg", "actors stopped", "err", err)
}
