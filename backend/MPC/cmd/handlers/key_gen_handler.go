package handlers

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
	"mpc/pkg/storage"
	"strings"
	"sync"
	"time"
)

type MemoryMpcServer struct {
	pb.UnimplementedWalletMpcServiceServer
	jobManager *dkg.JobManager
	router     network.Router
	shareStore storage.PartyShareStore

	preParamsMu   sync.Mutex
	preParams     []*keygen.LocalPreParams
	preParamsDone bool

	parties []*tss.PartyID
}

func NewMemoryMpcServer(jobManager *dkg.JobManager, router network.Router, shareStore storage.PartyShareStore) *MemoryMpcServer {
	return &MemoryMpcServer{
		jobManager: jobManager,
		router:     router,
		shareStore: shareStore,
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

func generateParams(idx int, preParamsList []*keygen.LocalPreParams, wg *sync.WaitGroup, errs chan<- error) {
	defer wg.Done()
	pp, err := keygen.GeneratePreParams(2 * time.Minute)

	if err != nil {
		errs <- fmt.Errorf("party %d: %w", idx, err)
		return
	}
	preParamsList[idx] = pp
}

func (s *MemoryMpcServer) generatePreParamsList(sortedIds []*tss.PartyID, jobID string) []*keygen.LocalPreParams {
	preParamsList := make([]*keygen.LocalPreParams, len(sortedIds))

	var wg sync.WaitGroup
	errs := make(chan error, len(sortedIds))

	for i := range preParamsList {
		wg.Add(1)
		go generateParams(i, preParamsList, &wg, errs)
	}

	wg.Wait()
	close(errs)

	if err := <-errs; err != nil {
		s.jobManager.FailJob(jobID, "failed to generate pre-params: "+err.Error())
		return nil
	}

	return preParamsList
}

func (s *MemoryMpcServer) cachedPreParams(sortedIds []*tss.PartyID, jobID string) []*keygen.LocalPreParams {
	s.preParamsMu.Lock()
	defer s.preParamsMu.Unlock()

	if s.preParamsDone {
		return s.preParams
	}

	list := s.generatePreParamsList(sortedIds, jobID)
	if list != nil {
		s.preParams = list
		s.preParamsDone = true
	}
	return list
}

func (s *MemoryMpcServer) GetOrGenerateSortedPartyIDs() (*tss.PartyID, []*tss.PartyID) {
	if len(s.parties) == 3 {
		return s.parties[0], s.parties
	}

	pID1 := tss.NewPartyID("1", "Node-1", big.NewInt(10))
	pID2 := tss.NewPartyID("2", "Node-2", big.NewInt(20))
	pID3 := tss.NewPartyID("3", "Node-3", big.NewInt(30))

	sortedIds := tss.SortPartyIDs([]*tss.PartyID{pID1, pID2, pID3})

	s.parties = sortedIds
	return pID1, sortedIds
}

func (s *MemoryMpcServer) executeKeyGen(jobID, walletID string, netParams *chaincfg.Params) {
	pID1, sortedIds := s.GetOrGenerateSortedPartyIDs()

	preParamsList := s.cachedPreParams(sortedIds, walletID)
	if preParamsList == nil {
		return
	}

	cfg := dkg.KeyGenConfig{
		WalletID:    walletID,
		PartyID:     pID1,
		AllPartyIDs: sortedIds,
		Threshold:   1,
		Router:      s.router,
		Network:     netParams,
		PreParams:   preParamsList,
		Timeout:     2 * time.Minute,
	}

	result, err := dkg.RunKeyGen(context.Background(), cfg)
	if err != nil {
		s.jobManager.FailJob(jobID, err.Error())
		return
	}

	if err := s.shareStore.Save(result.WalletId, result.SaveData); err != nil {
		s.jobManager.FailJob(jobID, "failed to persist data: "+err.Error())
		return
	}

	s.jobManager.CompleteJob(jobID, result)
}

func (s *MemoryMpcServer) StartKeyGen(ctx context.Context, req *pb.StartKeyGenRequest) (*pb.StartKeyGenResponse, error) {
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

func (s *MemoryMpcServer) GetKeyGenStatus(ctx context.Context, req *pb.GetKeyGenStatusRequest) (*pb.GetKeyGenStatusResponse, error) {
	if req.JobId == "" {
		return nil, status.Error(codes.InvalidArgument, "jobId is required")
	}

	job, found := s.jobManager.GetJob(req.JobId)
	if !found {
		return nil, status.Error(codes.NotFound, "job not found")
	}

	return &pb.GetKeyGenStatusResponse{
		Status:         job.Status,
		WalletId:       job.WalletID,
		PublicKeyHex:   job.PublicKeyHex,
		BitcoinAddress: job.BitcoinAddress,
		ErrorMessage:   job.ErrMsg,
		ChainCodeHex:   job.ChainCodeHex,
		Xpub:           job.XPub,
	}, nil
}
