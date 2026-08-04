package server

import (
	"context"
	"fmt"
	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/tss"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/big"
	pb "mpc/gen/mpc"
	"mpc/pkg/dkg"
	"mpc/pkg/network"
	"strings"
	"time"
)

type MpcServer struct {
	jobManager *dkg.JobManager
	router     network.Router
}

func NewMpcServer(jobManager *dkg.JobManager, router network.Router) *MpcServer {
	return &MpcServer{
		jobManager: jobManager,
		router:     router,
	}
}

func getNetwork(network string) (*chaincfg.Params, error) {
	switch strings.ToLower(strings.TrimSpace(network)) {
	case "mainnet", "livenet", "":
		return &chaincfg.MainNetParams, nil
	case "testnet", "testnet3":
		return &chaincfg.TestNet3Params, nil
	case "regtest":
		return &chaincfg.RegressionNetParams, nil
	case "simnet":
		return &chaincfg.SimNetParams, nil
	case "signet":
		return &chaincfg.SigNetParams, nil
	default:
		return nil, fmt.Errorf("unsupported bitcoin network: %q", network)
	}
}

func checkErrors(ctx context.Context, req *pb.StartKeyGenRequest) error {
	if err := ctx.Err(); err != nil {
		return status.Error(codes.Canceled, "request was canceled by client")
	}
	if req.WalletId == "" {
		return status.Error(codes.InvalidArgument, "Wallet id is required")
	}
	if req.IdempotencyKey == "" {
		return status.Error(codes.InvalidArgument, "idempotency_key is required")
	}

	return nil
}

func (s *MpcServer) executeKeyGen(jobID, walletID string, netParams *chaincfg.Params) {
	pID1 := tss.NewPartyID("1", "Node-1", big.NewInt(10))
	pID2 := tss.NewPartyID("2", "Node-2", big.NewInt(20))
	pID3 := tss.NewPartyID("3", "Node-3", big.NewInt(30))

	sortedIds := tss.SortedPartyIDs{pID1, pID2, pID3}
	preParams, err := keygen.GeneratePreParams(2 * time.Minute)
	if err != nil {
		s.jobManager.FailJob(jobID, "failed to generate pre-params: "+err.Error())
		return
	}

	cfg := dkg.KeyGenConfig{
		walletID,
		pID1,
		sortedIds,
		1,
		s.router,
		netParams,
		preParams,
		2 * time.Minute,
	}

	result, err := dkg.RunKeyGen(context.Background(), cfg)
	if err != nil {
		s.jobManager.FailJob(jobID, err.Error())
		return
	}

	s.jobManager.CompleteJob(jobID, result)
}

func (s *MpcServer) StartKeyGen(ctx context.Context, req *pb.StartKeyGenRequest) (*pb.StartKeyGenResponse, error) {
	err := checkErrors(ctx, req)
	if err != nil {
		return nil, err
	}

	netParams, err := getNetwork(req.Network)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	candidateJobID := uuid.New().String()
	jobID, isNewJob := s.jobManager.GetOrCreateJob(req.IdempotencyKey, candidateJobID, req.WalletId)

	if !isNewJob {
		return &pb.StartKeyGenResponse{JobId: jobID}, nil
	}

	go s.executeKeyGen(jobID, req.WalletId, netParams)

	return &pb.StartKeyGenResponse{JobId: jobID}, nil
}
