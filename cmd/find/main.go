// Program find looks up a user by ID at gRPC server and prints username if user was found.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/go-kit/log"
	"google.golang.org/grpc"

	"github.com/marselester/ddd-err/apiclient"
)

func main() {
	grpcAddr := flag.String("grpc", ":8080", "gRPC API address")
	userID := flag.String("user-id", "", "user ID to look for")
	flag.Parse()

	exitCode := 1
	defer func() { os.Exit(exitCode) }()

	var logger log.Logger
	{
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	if *userID == "" {
		logger.Log("msg", "user ID is required")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		cancel()
	}()

	conn, err := grpc.DialContext(ctx, *grpcAddr, grpc.WithInsecure())
	if err != nil {
		logger.Log("msg", "grpc dial", "err", err)
		return
	}
	defer conn.Close()

	svc := apiclient.NewGRPCUserClient(conn)
	user, err := svc.FindUserByID(ctx, *userID)
	if err != nil {
		logger.Log("msg", "user search", "err", err)
		return
	}

	fmt.Println(user.Username)

	exitCode = 0
}
