package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mpc/cmd/server"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"

	pb "mpc/gen/mpc"
	"mpc/pkg/cron"
	"mpc/pkg/dkg"
	"mpc/pkg/kms"
	"mpc/pkg/network"
	"mpc/pkg/storage"
)

type AppConfig struct {
	Server struct {
		Port string `json:"port"`
	} `json:"server"`
	KMS struct {
		ConfigFile string `json:"configFile"`
	} `json:"kms"`
	Storage struct {
		ShareStoreDir string `json:"shareStoreDir"`
	} `json:"storage"`
	Rotation struct {
		Enabled bool `json:"enabled"`
	} `json:"rotation"`
}

func loadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config json: %w", err)
	}

	if cfg.Server.Port == "" {
		cfg.Server.Port = ":50051"
	}

	return &cfg, nil
}

func main() {
	configPath := "config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load configuration from %s: %v", configPath, err)
	}

	if cfg.KMS.ConfigFile == "" {
		log.Fatal("kms.configFile is required in config.json")
	}
	if cfg.Storage.ShareStoreDir == "" {
		log.Fatal("storage.shareStoreDir is required in config.json")
	}

	jobManager := dkg.NewJobManager()
	router := network.NewMemoryRouter()

	localKMS, err := kms.NewLocalFileKMS(cfg.KMS.ConfigFile)
	if err != nil {
		log.Fatalf("failed to load KMS config: %v", err)
	}

	shareStore, err := storage.NewEncryptedLocalShareStore(cfg.Storage.ShareStoreDir, localKMS)
	if err != nil {
		log.Fatalf("failed to init share store: %v", err)
	}

	rotationMgr := cron.NewKeyRotationManager(localKMS, shareStore)

	if err := rotationMgr.RecoverOnStartup(); err != nil {
		log.Fatalf("failed to recover key rotation: %v", err)
	}

	if cfg.Rotation.Enabled {
		if err := rotationMgr.StartQuarterlyRotation(); err != nil {
			log.Fatalf("failed to start key rotation cron: %v", err)
		}
	}

	mpcServer := server.NewMemoryMpcServer(jobManager, router, shareStore, 1)

	lis, err := net.Listen("tcp", cfg.Server.Port)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", cfg.Server.Port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWalletMpcServiceServer(grpcServer, mpcServer.KeyGenHandler)
	pb.RegisterTransactionMpcServiceServer(grpcServer, mpcServer.TransactionHandler)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		<-stop
		log.Println("shutting down...")

		rotationMgr.Stop()
		grpcServer.GracefulStop()
	}()

	log.Printf("MPC gRPC server listening on %s\n", cfg.Server.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}
