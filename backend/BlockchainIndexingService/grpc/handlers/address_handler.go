package handlers

import (
	"BlockchainIndexingService/errors"
	pb "BlockchainIndexingService/proto"
	"BlockchainIndexingService/services"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AddressGRPCHandler struct {
	pb.UnimplementedAddressServiceServer
	service *services.AddressService
}

func NewAddressGRPCHandler(service *services.AddressService) *AddressGRPCHandler {
	return &AddressGRPCHandler{service: service}
}

func (handler *AddressGRPCHandler) GetAddressUTXOs(ctx context.Context, req *pb.GetAddressUTXOsRequest) (*pb.GetAddressUTXOsResponse, error) {
	if req.GetAddress() == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}

	utxos, err := handler.service.GetAddressUTXOs(ctx, req.GetAddress())
	if err != nil {
		return nil, errors.MapToGRPCError(err)
	}

	pbUTXOs := make([]*pb.UTXO, len(utxos))
	for i, u := range utxos {
		pbUTXOs[i] = &pb.UTXO{
			TxId:          u.TxID,
			Vout:          u.Vout,
			Address:       u.Address,
			AmountSats:    u.AmountSats,
			BlockHeight:   u.BlockHeight,
			Confirmations: u.Confirmations,
		}
	}

	return &pb.GetAddressUTXOsResponse{Utxos: pbUTXOs}, nil
}

func (handler *AddressGRPCHandler) GetWalletBalance(ctx context.Context, req *pb.GetWalletBalanceRequest) (*pb.GetWalletBalanceResponse, error) {
	if req.GetAddress() == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}

	balance, err := handler.service.GetWalletBalance(ctx, req.GetAddress())
	if err != nil {
		return nil, errors.MapToGRPCError(err)
	}

	return &pb.GetWalletBalanceResponse{ConfirmedBalance: balance}, nil
}
