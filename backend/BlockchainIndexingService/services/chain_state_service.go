package services

import (
	"BlockchainIndexingService/dtos"
	"BlockchainIndexingService/errors"
	"BlockchainIndexingService/repository"
	"context"
	"github.com/btcsuite/btcd/chainhash/v2"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire/v2"
)

type ChainStateService struct {
	rpcClient *rpcclient.Client
	store     repository.Store
}

func NewChainStateService(rpcClient *rpcclient.Client, store repository.Store) *ChainStateService {
	return &ChainStateService{rpcClient: rpcClient, store: store}
}

func (service *ChainStateService) GetBlockByHash(ctx context.Context, hash string) (*wire.MsgBlock, error) {
	blockHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, errors.ErrInvalidArgument
	}

	block, err := service.rpcClient.GetBlock(blockHash)
	if err != nil {
		return nil, errors.ErrInternalServer
	}

	return block, nil

}

func (service *ChainStateService) GetBlockByHeight(ctx context.Context, height int) (*wire.MsgBlock, error) {
	blockHash, err := service.rpcClient.GetBlockHash(int64(height))
	if err != nil {
		return nil, errors.ErrInvalidArgument
	}

	block, err := service.rpcClient.GetBlock(blockHash)
	if err != nil {
		return nil, errors.ErrInternalServer
	}

	return block, nil
}

func (service *ChainStateService) GetChainStatus(ctx context.Context) (dtos.ChainStatusResponse, error) {
	info, err := service.rpcClient.GetBlockChainInfo()
	if err != nil {
		return dtos.ChainStatusResponse{}, errors.ErrInternalServer
	}

	indexedHeight, err := service.store.LastIndexedHeight(ctx)
	if err != nil {
		return dtos.ChainStatusResponse{}, errors.ErrInternalServer
	}

	var syncPct float64
	if info.Blocks > 0 {
		syncPct = float64(indexedHeight) / float64(info.Blocks)
		if syncPct > 1.0 {
			syncPct = 1.0
		}
	} else if indexedHeight >= 0 {
		syncPct = 1.0
	}

	res := dtos.ChainStatusResponse{
		Chain:                info.Chain,
		NodeHeight:           int64(info.Blocks),
		IndexedHeight:        indexedHeight,
		SyncPercentage:       syncPct,
		BestBlockHash:        info.BestBlockHash,
		VerificationProgress: info.VerificationProgress,
	}

	return res, nil
}

func (service *ChainStateService) GetFeeEstimation(ctx context.Context, targetBlocks int64) (dtos.FeeEstimationResponse, error) {

	feeResult, err := service.rpcClient.EstimateSmartFee(targetBlocks, nil)
	if err != nil {
		return dtos.FeeEstimationResponse{}, errors.ErrInternalServer
	}

	var btcPerKb float64
	if feeResult.FeeRate != nil {
		btcPerKb = *feeResult.FeeRate
	}

	satPerVByte := btcPerKb * 100000.0

	res := dtos.FeeEstimationResponse{
		TargetBlocks: targetBlocks,
		FeeRateSatVB: satPerVByte,
		FeeRateBtcKB: btcPerKb,
	}

	return res, nil

}
