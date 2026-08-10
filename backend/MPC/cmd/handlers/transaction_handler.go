package handlers

import (
	"context"
	"fmt"
	pb "mpc/gen/mpc"
	"mpc/pkg/network"
	"mpc/pkg/signer"
	"mpc/pkg/storage"
	"time"

	"github.com/bnb-chain/tss-lib/v2/tss"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TransactionHandler interface {
	pb.TransactionMpcServiceServer
}

type MemoryTransactionHandler struct {
	pb.UnimplementedTransactionMpcServiceServer

	signer     signer.Signer
	router     network.Router
	shareStore storage.PartyShareStore

	ownPartyID  *tss.PartyID
	allPartyIDs []*tss.PartyID
}

func NewMemoryTransactionHandler(signer signer.Signer, router network.Router, shareStore storage.PartyShareStore, ownPartyID *tss.PartyID, allPartyIDs []*tss.PartyID) TransactionHandler {
	return &MemoryTransactionHandler{
		signer:      signer,
		router:      router,
		shareStore:  shareStore,
		ownPartyID:  ownPartyID,
		allPartyIDs: allPartyIDs,
	}
}

func checkSignRequest(ctx context.Context, req *pb.SignRequest) error {
	if err := ctx.Err(); err != nil {
		return status.Error(codes.Canceled, "request was canceled by client")
	}
	if req.WalletId == "" {
		return status.Error(codes.InvalidArgument, "wallet_id is required")
	}
	if req.IdempotencyKey == "" {
		return status.Error(codes.InvalidArgument, "idempotency_key is required")
	}
	if len(req.TxHash) == 0 {
		return status.Error(codes.InvalidArgument, "tx_hash is required")
	}

	return nil
}

func (h *MemoryTransactionHandler) Sign(ctx context.Context, req *pb.SignRequest) (*pb.SignResponse, error) {
	if err := checkSignRequest(ctx, req); err != nil {
		return nil, err
	}

	if _, err := GetNetwork(req.Network); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !h.shareStore.Exists(req.WalletId) {
		return nil, status.Error(codes.NotFound, "no key share found for wallet")
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	signature, err := h.signer.Sign(ctx, req.WalletId, req.TxHash)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("signing failed: %v", err))
	}

	return &pb.SignResponse{
		WalletId:   req.WalletId,
		Signature:  signature.Signature,
		R:          signature.R,
		S:          signature.S,
		RecoveryId: signature.RecoveryID,
	}, nil
}
