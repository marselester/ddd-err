// Program find looks up a user by ID at gRPC server and prints username if user was found.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"google.golang.org/grpc"

	"github.com/marselester/ddd-err/apiclient"
)

func main() {
	grpcAddr := flag.String("grpc", ":8080", "gRPC API address")
	userID := flag.String("user-id", "", "user ID to look for")
	flag.Parse()
	if *userID == "" {
		log.Fatal("user ID is required")
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
		log.Fatalf("grpc dial failed: %v", err)
	}
	defer conn.Close()

	svc := apiclient.NewGRPCUserClient(conn)
	user, err := svc.FindUserByID(ctx, *userID)
	if err != nil {
		log.Fatalf("user search failed: %v", err)
	}

	fmt.Println(user.Username)
}
