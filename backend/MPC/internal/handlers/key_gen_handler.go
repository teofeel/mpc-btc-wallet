package handlers

import (
	"context"
	"fmt"
	pb "mpc/gen/mpc"
	"mpc/pkg/dkg"
	"mpc/pkg/network"
	"mpc/pkg/storage"
	"mpc/pkg/utils"
	"sync"
	"time"

	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/tss"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type KeyGenHandler interface {
	pb.WalletMpcServiceServer
}

type MemoryKeyGenHandler struct {
	pb.UnimplementedWalletMpcServiceServer

	jobManager *dkg.JobManager
	router     network.Router
	shareStore storage.PartyShareStore

	ownPartyID  *tss.PartyID
	allPartyIDs []*tss.PartyID

	preParamsMu   sync.Mutex
	preParams     []*keygen.LocalPreParams
	preParamsDone bool
}

func NewMemoryKeyGenHandler(jobManager *dkg.JobManager, router network.Router, shareStore storage.PartyShareStore, ownPartyID *tss.PartyID, allPartyIDs []*tss.PartyID) KeyGenHandler {
	return &MemoryKeyGenHandler{
		jobManager:  jobManager,
		router:      router,
		shareStore:  shareStore,
		ownPartyID:  ownPartyID,
		allPartyIDs: allPartyIDs,
	}
}

func checkKeyGenRequest(ctx context.Context, req *pb.StartKeyGenRequest) error {
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

func (h *MemoryKeyGenHandler) generatePreParamsList(sortedIds []*tss.PartyID, jobID string) []*keygen.LocalPreParams {
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
		h.jobManager.FailJob(jobID, "failed to generate pre-params: "+err.Error())
		return nil
	}

	return preParamsList
}

func (h *MemoryKeyGenHandler) cachedPreParams(sortedIds []*tss.PartyID, jobID string) []*keygen.LocalPreParams {
	h.preParamsMu.Lock()
	defer h.preParamsMu.Unlock()

	if h.preParamsDone {
		return h.preParams
	}

	list := h.generatePreParamsList(sortedIds, jobID)
	if list != nil {
		h.preParams = list
		h.preParamsDone = true
	}
	return list
}

func (h *MemoryKeyGenHandler) executeKeyGen(jobID, walletID string, netParams *chaincfg.Params) {
	preParamsList := h.cachedPreParams(h.allPartyIDs, walletID)
	if preParamsList == nil {
		return
	}

	cfg := dkg.KeyGenConfig{
		WalletID:    walletID,
		PartyID:     h.ownPartyID,
		AllPartyIDs: h.allPartyIDs,
		Threshold:   1,
		Router:      h.router,
		Network:     netParams,
		PreParams:   preParamsList,
		Timeout:     2 * time.Minute,
	}

	result, err := dkg.RunKeyGen(context.Background(), cfg)
	if err != nil {
		h.jobManager.FailJob(jobID, err.Error())
		return
	}

	if err := h.shareStore.Save(result.WalletId, result.SaveData); err != nil {
		h.jobManager.FailJob(jobID, "failed to persist data: "+err.Error())
		return
	}

	if err := h.shareStore.SaveAll(result.WalletId, result.AllSaveData); err != nil {
		h.jobManager.FailJob(jobID, "failed to persist party shares: "+err.Error())
		return
	}

	h.jobManager.CompleteJob(jobID, result)
}

func (h *MemoryKeyGenHandler) StartKeyGen(ctx context.Context, req *pb.StartKeyGenRequest) (*pb.StartKeyGenResponse, error) {
	err := checkKeyGenRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	netParams, err := utils.GetNetwork(req.Network)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	candidateJobID := uuid.New().String()
	jobID, isNewJob := h.jobManager.GetOrCreateJob(req.IdempotencyKey, candidateJobID, req.WalletId)

	if !isNewJob {
		return &pb.StartKeyGenResponse{JobId: jobID}, nil
	}

	go h.executeKeyGen(jobID, req.WalletId, netParams)

	return &pb.StartKeyGenResponse{JobId: jobID}, nil
}

func (h *MemoryKeyGenHandler) GetKeyGenStatus(ctx context.Context, req *pb.GetKeyGenStatusRequest) (*pb.GetKeyGenStatusResponse, error) {
	if req.JobId == "" {
		return nil, status.Error(codes.InvalidArgument, "jobId is required")
	}

	job, found := h.jobManager.GetJob(req.JobId)
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
