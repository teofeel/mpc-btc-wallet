package services

import (
	"BlockchainIndexingService/repository"
	"BlockchainIndexingService/utils"
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcd/chainhash/v2"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type TransactionService struct {
	rpcClient *rpcclient.Client
	store     repository.Store
}

func NewTransactionService(rpcClient *rpcclient.Client, store repository.Store) *TransactionService {
	return &TransactionService{rpcClient: rpcClient, store: store}
}

func (service *TransactionService) GetTransaction(w http.ResponseWriter, r *http.Request) {
	txID := chi.URLParam(r, "txID")
	hash, err := chainhash.NewHashFromStr(txID)
	if err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	tx, err := service.rpcClient.GetTransaction(hash)
	if err != nil {
		utils.WriteErr(w, http.StatusNotFound, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, tx)
}

func (service *TransactionService) BroadcastTransaction(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Hex string
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	raw, err := hex.DecodeString(body.Hex)
	if err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	msgTx, err := utils.DecodeTxBytes(raw)
	if err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	txHash, err := service.rpcClient.SendRawTransaction(msgTx, false)
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"txID": txHash.String()})
}

func (service *TransactionService) DecodeTransaction(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Hex string
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	raw, err := hex.DecodeString(body.Hex)
	if err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := service.rpcClient.DecodeRawTransaction(raw)
	if err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, result)
}
