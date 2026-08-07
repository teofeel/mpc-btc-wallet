package handlers

import (
	"BlockchainIndexingService/errors"
	pb "BlockchainIndexingService/proto"
	"BlockchainIndexingService/services"
	"bytes"
	"context"
	"github.com/btcsuite/btcd/wire/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChainStateGRPCHandler struct {
	pb.UnimplementedChainStateServiceServer
	service *services.ChainStateService
}

func NewChainStateGRPCHandler(service *services.ChainStateService) *ChainStateGRPCHandler {
	return &ChainStateGRPCHandler{service: service}
}

func (handler *ChainStateGRPCHandler) GetBlockByHash(ctx context.Context, req *pb.GetBlockByHashRequest) (*pb.GetBlockByHashResponse, error) {
	if req.GetHash() == "" {
		return nil, status.Error(codes.InvalidArgument, "block hash is required")
	}

	block, err := handler.service.GetBlockByHash(ctx, req.GetHash())
	if err != nil {
		return nil, errors.MapToGRPCError(err)
	}

	var buf bytes.Buffer
	if err := block.BtcEncode(&buf, wire.ProtocolVersion, wire.WitnessEncoding); err != nil {
		return nil, status.Error(codes.Internal, "failed to serialize block")
	}

	return &pb.GetBlockByHashResponse{RawBlock: buf.Bytes()}, nil
}

func (handler *ChainStateGRPCHandler) GetBlockByHeight(ctx context.Context, req *pb.GetBlockByHeightRequest) (*pb.GetBlockByHeightResponse, error) {
	if req.GetHeight() < 0 {
		return nil, status.Error(codes.InvalidArgument, "height must be non negative")
	}

	block, err := handler.service.GetBlockByHeight(ctx, int(req.GetHeight()))
	if err != nil {
		return nil, errors.MapToGRPCError(err)
	}

	var buf bytes.Buffer
	if err := block.BtcEncode(&buf, wire.ProtocolVersion, wire.WitnessEncoding); err != nil {
		return nil, status.Error(codes.Internal, "failed to serialize block")
	}

	return &pb.GetBlockByHeightResponse{RawBlock: buf.Bytes()}, nil
}

func (handler *ChainStateGRPCHandler) GetChainStatus(ctx context.Context, req *pb.GetChainStatusRequest) (*pb.GetChainStatusResponse, error) {
	res, err := handler.service.GetChainStatus(ctx)
	if err != nil {
		return nil, errors.MapToGRPCError(err)
	}

	return &pb.GetChainStatusResponse{
		Status: &pb.ChainStatusResponseDto{
			Chain:                res.Chain,
			NodeHeight:           res.NodeHeight,
			IndexedHeight:        res.IndexedHeight,
			SyncPercentage:       res.SyncPercentage,
			BestBlockHash:        res.BestBlockHash,
			VerificationProgress: res.VerificationProgress,
		},
	}, nil
}

func (handler *ChainStateGRPCHandler) GetFeeEstimation(ctx context.Context, req *pb.GetFeeEstimationRequest) (*pb.GetFeeEstimationResponse, error) {
	if req.GetTargetBlocks() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "target blocks must be greater than zero")
	}

	res, err := handler.service.GetFeeEstimation(ctx, req.GetTargetBlocks())
	if err != nil {
		return nil, errors.MapToGRPCError(err)
	}

	return &pb.GetFeeEstimationResponse{
		Fee: &pb.FeeEstimationResponseDto{
			TargetBlocks: res.TargetBlocks,
			FeeRateSatVb: res.FeeRateSatVB,
			FeeRateBtcKb: res.FeeRateBtcKB,
		},
	}, nil
}
