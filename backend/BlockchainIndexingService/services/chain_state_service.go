package services

import (
	"BlockchainIndexingService/dtos"
	"BlockchainIndexingService/repository"
	"BlockchainIndexingService/utils"
	"github.com/btcsuite/btcd/chainhash/v2"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type ChainStateService struct {
	rpcClient *rpcclient.Client
	store     repository.Store
}

func NewChainStateService(rpcClient *rpcclient.Client, store repository.Store) *ChainStateService {
	return &ChainStateService{rpcClient: rpcClient, store: store}
}

func (service *ChainStateService) GetBlockByHash(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	blockHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	block, err := service.rpcClient.GetBlock(blockHash)
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, block)
}

func (service *ChainStateService) GetBlockByHeight(w http.ResponseWriter, r *http.Request) {
	height := chi.URLParam(r, "height")

	heightInt, err := strconv.Atoi(height)
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	blockHash, err := service.rpcClient.GetBlockHash(int64(heightInt))
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	block, err := service.rpcClient.GetBlock(blockHash)
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, block)
}

func (service *ChainStateService) GetChainStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	info, err := service.rpcClient.GetBlockChainInfo()
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	indexedHeight, err := service.store.LastIndexedHeight(ctx)
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	var syncPct float64
	if info.Blocks > 0 {
		syncPct = float64(info.Blocks) / float64(indexedHeight)
		if syncPct > 100 {
			syncPct = 100
		}
	} else if indexedHeight >= 0 {
		syncPct = 100
	}

	res := dtos.ChainStatusResponse{
		Chain:                info.Chain,
		NodeHeight:           int64(info.Blocks),
		IndexedHeight:        indexedHeight,
		SyncPercentage:       syncPct,
		BestBlockHash:        info.BestBlockHash,
		VerificationProgress: info.VerificationProgress,
	}

	utils.WriteJSON(w, http.StatusOK, res)
}

func (service *ChainStateService) GetFeeEstimation(w http.ResponseWriter, r *http.Request) {
	targetBlocksStr := r.URL.Query().Get("blocks")
	targetBlocks := int64(6)

	if targetBlocksStr != "" {
		parsedTarget, err := strconv.ParseInt(targetBlocksStr, 10, 64)

		if err != nil || parsedTarget <= 0 {
			utils.WriteErr(w, http.StatusBadRequest, "invalid 'blocks' parameter; must be a positive integer")
			return
		}

		targetBlocks = parsedTarget
	}

	feeResult, err := service.rpcClient.EstimateSmartFee(targetBlocks, nil)
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
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

	utils.WriteJSON(w, http.StatusOK, res)
}
