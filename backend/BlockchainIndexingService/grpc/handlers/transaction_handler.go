package handlers

import (
	"BlockchainIndexingService/errors"
	pb "BlockchainIndexingService/proto"
	"BlockchainIndexingService/services"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TransactionGRPCHandler struct {
	pb.UnimplementedTransactionServiceServer
	service *services.TransactionService
}

func NewTransactionGRPCHandler(service *services.TransactionService) *TransactionGRPCHandler {
	return &TransactionGRPCHandler{service: service}
}

func (h *TransactionGRPCHandler) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.GetTransactionResponse, error) {
	if req.GetTxId() == "" {
		return nil, status.Error(codes.InvalidArgument, "txID is required")
	}

	tx, err := h.service.GetTransaction(ctx, req.GetTxId())
	if err != nil {
		return nil, errors.MapToGRPCError(err)
	}

	return &pb.GetTransactionResponse{
		Transaction: &pb.TransactionResponseDto{
			TxId:          tx.TxID,
			AmountSats:    int64(tx.Amount * 1e8),
			Confirmations: tx.Confirmations,
			BlockHash:     tx.BlockHash,
			BlockTime:     tx.BlockTime,
			Hex:           tx.Hex,
		},
	}, nil
}

func (h *TransactionGRPCHandler) BroadcastTransaction(ctx context.Context, req *pb.BroadcastTransactionRequest) (*pb.BroadcastTransactionResponse, error) {
	if req.GetHexTx() == "" {
		return nil, status.Error(codes.InvalidArgument, "hexTX is required")
	}

	txID, err := h.service.BroadcastTransaction(ctx, req.GetHexTx())
	if err != nil {
		return nil, errors.MapToGRPCError(err)
	}

	return &pb.BroadcastTransactionResponse{TxId: txID}, nil
}

func (h *TransactionGRPCHandler) DecodeTransaction(ctx context.Context, req *pb.DecodeTransactionRequest) (*pb.DecodeTransactionResponse, error) {
	if req.GetHexTx() == "" {
		return nil, status.Error(codes.InvalidArgument, "hexTX is required")
	}

	result, err := h.service.DecodeTransaction(ctx, req.GetHexTx())
	if err != nil {
		return nil, errors.MapToGRPCError(err)
	}

	return &pb.DecodeTransactionResponse{
		Transaction: &pb.DecodedTransactionResponseDto{
			TxId:     result.Txid,
			Version:  result.Version,
			LockTime: result.LockTime,
			Size:     uint32(result.Size),
			Vsize:    uint32(result.Vsize),
			Weight:   uint32(result.Weight),
		},
	}, nil
}
