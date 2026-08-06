package services

import (
	"BlockchainIndexingService/repository"
	"BlockchainIndexingService/utils"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type AddressService struct {
	rpcClient *rpcclient.Client
	store     repository.Store
}

func NewAddressService(rpcClient *rpcclient.Client, store repository.Store) *AddressService {
	return &AddressService{rpcClient: rpcClient, store: store}
}

func (service *AddressService) GetAddressUTXOs(w http.ResponseWriter, r *http.Request) {
	addr := chi.URLParam(r, "address")

	tip, err := service.rpcClient.GetBlockCount()
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	utxos, err := service.store.UTXOsForAddress(r.Context(), addr, tip)
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, utxos)
}

func (service *AddressService) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	addr := chi.URLParam(r, "address")

	balance, err := service.store.BalanceForAddress(r.Context(), addr)
	if err != nil {
		utils.WriteErr(w, http.StatusNotFound, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]int64{"confirmed_balance": balance})
}

func (service *AddressService) GetWalletTransactions(w http.ResponseWriter, r *http.Request) {
	utils.WriteErr(w, http.StatusNotImplemented, "not yet implemented")
}
