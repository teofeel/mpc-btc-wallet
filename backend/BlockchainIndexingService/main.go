package main

import (
	"BlockchainIndexingService/controllers"
	"BlockchainIndexingService/db"
	"BlockchainIndexingService/eureka"
	"BlockchainIndexingService/grpc/handlers"
	"BlockchainIndexingService/indexers"
	pb "BlockchainIndexingService/proto"
	"BlockchainIndexingService/repository"
	"BlockchainIndexingService/services"
	"BlockchainIndexingService/utils"
	"context"
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func CreateRpcClient() (*rpcclient.Client, error) {

	rpcHost := utils.GetEnv("BTC_RPC_HOST", "127.0.0.1:18443")
	rpcUser := utils.GetEnv("BTC_RPC_USER", "user")
	rpcPass := utils.GetEnv("BTC_RPC_PASS", "pass")

	connCfg := &rpcclient.ConnConfig{
		Host:         rpcHost,
		User:         rpcUser,
		Pass:         rpcPass,
		HTTPPostMode: true,
		DisableTLS:   true,
		Params:       "regtest",
	}

	return rpcclient.New(connCfg, nil)

}

func StartRpcServer(addressService *services.AddressService, chainStateService *services.ChainStateService, transactionService *services.TransactionService) *grpc.Server {
	grpcPort := utils.GetEnv("GRPC_PORT", ":50052")

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	addressGRPCHandler := handlers.NewAddressGRPCHandler(addressService)
	chainStateGRPCHandler := handlers.NewChainStateGRPCHandler(chainStateService)
	transactionGRPCHandler := handlers.NewTransactionGRPCHandler(transactionService)

	pb.RegisterAddressServiceServer(grpcServer, addressGRPCHandler)
	pb.RegisterChainStateServiceServer(grpcServer, chainStateGRPCHandler)
	pb.RegisterTransactionServiceServer(grpcServer, transactionGRPCHandler)

	go func() {
		log.Printf("gRPC server listening on %s\n", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	return grpcServer
}

func StartHTTPServer(addressService *services.AddressService, chainStateService *services.ChainStateService, transactionService *services.TransactionService) *http.Server {
	router := chi.NewRouter()

	router.Route("/api/v1/blockchain", func(r chi.Router) {
		addressController := controllers.NewAddressController(addressService)
		addressController.RegisterRoutes(r)

		transactionController := controllers.NewTransactionController(transactionService)
		transactionController.RegisterRoutes(r)

		chainStateController := controllers.NewChainStateController(chainStateService)
		chainStateController.RegisterRoutes(r)
	})

	port := utils.GetEnv("PORT", ":8083")
	httpServer := &http.Server{
		Addr:    port,
		Handler: router,
	}

	go func() {
		fmt.Printf("Server running on %s\n", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	return httpServer
}

func main() {
	ctx, cancelApp := context.WithCancel(context.Background())
	defer cancelApp()

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startupCancel()

	eurCfg := eureka.DefaultConfig()
	cleanup, err := eureka.RegisterWithEureka(startupCtx, eurCfg)
	if err != nil {
		log.Printf("Eureka registration failed, continuing without it: %v", err)
	} else {
		defer cleanup()
	}

	rpcClient, err := CreateRpcClient()
	if err != nil {
		log.Fatal(err)
	}
	defer rpcClient.Shutdown()

	info, err := rpcClient.GetBlockChainInfo()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("connecteed to bitcoind: chain=%s height=%d", info.Chain, info.Blocks)

	pool, err := db.NewPostgresPool(ctx)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	store := repository.NewPostgresStore(pool)
	err = store.InitSchema(ctx)
	if err != nil {
		log.Fatalf("failed to initialize schema: %v", err)
	}

	log.Println("successfully connected to postgres")

	addressService := services.NewAddressService(rpcClient, store)
	transactionService := services.NewTransactionService(rpcClient, store)
	chainStateService := services.NewChainStateService(rpcClient, store)

	idx := indexers.NewIndexer(rpcClient, store)
	go idx.Run(ctx)

	grpcServer := StartRpcServer(addressService, chainStateService, transactionService)
	httpServer := StartHTTPServer(addressService, chainStateService, transactionService)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down servers...")

	cancelApp()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	grpcServer.GracefulStop()
	_ = httpServer.Shutdown(shutdownCtx)
}
