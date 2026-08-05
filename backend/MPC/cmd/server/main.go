package main

import (
	"log"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"

	"mpc/cmd/handlers"
	pb "mpc/gen/mpc"
	"mpc/pkg/dkg"
	"mpc/pkg/network"
)

func main() {
	jobManager := dkg.NewJobManager()
	router := network.NewMemoryRouter()
	mpcServer := handlers.NewMemoryMpcServer(jobManager, router)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen on :50051: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWalletMpcServiceServer(grpcServer, mpcServer)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		<-stop
		log.Println("shutting down...")
		grpcServer.GracefulStop()
	}()

	log.Println("MPC gRPC handlers listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC handlers failed: %v", err)
	}
}
