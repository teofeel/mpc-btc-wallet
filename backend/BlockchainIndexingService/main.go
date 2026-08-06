package main

import (
	"BlockchainIndexingService/controllers"
	"BlockchainIndexingService/db"
	"BlockchainIndexingService/indexers"
	"BlockchainIndexingService/repository"
	"BlockchainIndexingService/services"
	"BlockchainIndexingService/utils"
	"context"
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
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

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

	idx := indexers.NewIndexer(rpcClient, store)
	go idx.Run(ctx)

	router := chi.NewRouter()

	addressService := services.NewAddressService(rpcClient, store)
	addressController := controllers.NewAddressController(addressService)
	addressController.RegisterRoutes(router)

	transactionController := controllers.NewTransactionController()
	transactionController.RegisterRoutes(router)

	chainStateService := services.NewChainStateService(rpcClient, store)
	chainStateController := controllers.NewChainStateController(chainStateService)
	chainStateController.RegisterRoutes(router)

	port := utils.GetEnv("PORT", ":8082")

	fmt.Printf("Server running on %s\n", port)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatal(err)
	}
}
